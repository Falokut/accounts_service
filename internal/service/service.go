package service

import (
	"context"
	"time"

	"github.com/Falokut/accounts_service/internal/config"
	"github.com/Falokut/accounts_service/internal/events"
	"github.com/Falokut/accounts_service/internal/models"
	"github.com/Falokut/accounts_service/internal/repository"
	accounts_service "github.com/Falokut/accounts_service/pkg/accounts_service/v1/protos"
	"github.com/Falokut/accounts_service/pkg/jwt"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AccountsService interface {
	CreateAccount(ctx context.Context, dto models.CreateAccountDTO) error
	DeleteAccount(ctx context.Context, sessionID, machineID string) error
	RequestAccountVerificationToken(ctx context.Context, email, callbackURL string) error
	VerifyAccount(ctx context.Context, token string) error
	SignIn(ctx context.Context, dto models.SignInDTO) (sessionID string, err error)
	GetAccountID(ctx context.Context, sessionID, machineID string) (accountID string, err error)
	Logout(ctx context.Context, sessionID, machineID string) error
	RequestChangePasswordToken(ctx context.Context, email, callbackURL string) error
	ChangePassword(ctx context.Context, token, newPassword string) error
	GetAllSessions(ctx context.Context, sessionID, machineID string) (map[string]*models.SessionInfo, error)
	TerminateSessions(ctx context.Context, sessionID, machineID string, sessionsToTerminateIds []string) error
}

type AccountsServiceConfig struct {
	ChangePasswordTokenTTL             time.Duration
	ChangePasswordTokenSecret          string
	VerifyAccountTokenTTL              time.Duration
	VerifyAccountTokenSecret           string
	NumRetriesForTerminateSessions     uint32
	RetrySleepTimeForTerminateSessions time.Duration
	NonActivatedAccountTTL             time.Duration
	BcryptCost                         int
	SessionTTL                         time.Duration
}

type accountsService struct {
	accounts_service.UnimplementedAccountsServiceV1Server
	accountsRepository     repository.AccountRepository
	registrationRepository repository.RegistrationRepository
	sessionsRepository     repository.SessionsRepository
	logger                 *logrus.Logger
	cfg                    AccountsServiceConfig
	accountEvents          events.AccountsEventsMQ
	tokenDeliveryMQ        events.TokensDeliveryMQ
}

func NewAccountsService(repo repository.AccountRepository,
	logger *logrus.Logger,
	registrationRepository repository.RegistrationRepository,
	sessionsRepository repository.SessionsRepository,
	accountEvents events.AccountsEventsMQ,
	tokenDeliveryMQ events.TokensDeliveryMQ,
	cfg *AccountsServiceConfig) *accountsService {
	return &accountsService{accountsRepository: repo,
		logger:                 logger,
		registrationRepository: registrationRepository,
		sessionsRepository:     sessionsRepository,
		cfg:                    *cfg,
		tokenDeliveryMQ:        tokenDeliveryMQ,
		accountEvents:          accountEvents,
	}
}

func (s *accountsService) CreateAccount(ctx context.Context,
	dto models.CreateAccountDTO) (err error) {
	exist, err := s.accountsRepository.IsAccountWithEmailExist(ctx, dto.Email)
	if err != nil {
		return err
	}
	if exist {
		return models.Error(models.Conflict, "a user with this email address already exists. "+
			"please try another one or simple log in")
	}

	inCache, err := s.registrationRepository.IsAccountExist(ctx, dto.Email)
	if err != nil {
		return err
	}
	if inCache {
		return models.Error(models.Conflict, "a user with this email address already exists. "+
			"please try another one or verify email and log in")
	}

	s.logger.Info("Generating hash from password")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), s.cfg.BcryptCost)
	if err != nil {
		return models.Error(models.Internal, "can't generate password hash")
	}

	err = s.registrationRepository.SetAccount(ctx, dto.Email,
		models.RegisteredAccount{
			Username: dto.Username,
			Password: string(passwordHash),
		},
		s.cfg.NonActivatedAccountTTL)

	return err
}

func (s *accountsService) RequestAccountVerificationToken(ctx context.Context,
	email, callbackURL string) (err error) {
	exist, err := s.accountsRepository.IsAccountWithEmailExist(ctx, email)
	if err != nil {
		return err
	}
	if exist {
		return models.Error(models.InvalidArgument, "account already activated")
	}

	inCache, err := s.registrationRepository.IsAccountExist(ctx, email)
	if err != nil {
		return err
	}
	if !inCache {
		return models.Error(models.NotFound, "a account with this email address not exist")
	}

	token, err := jwt.GenerateToken(email, s.cfg.VerifyAccountTokenSecret, s.cfg.VerifyAccountTokenTTL)
	if err != nil {
		return err
	}

	err = s.tokenDeliveryMQ.RequestEmailVerificationTokenDelivery(ctx, email, token, callbackURL, s.cfg.VerifyAccountTokenTTL)
	return err
}

func (s *accountsService) VerifyAccount(ctx context.Context, token string) (err error) {
	s.logger.Info("Parsing token")
	email, err := jwt.ParseToken(token, config.GetConfig().JWT.VerifyAccountToken.Secret)
	if err != nil {
		return models.Error(models.InvalidArgument, err.Error())
	}

	err = s.createAccount(ctx, email)
	return err
}

func (s *accountsService) createAccount(ctx context.Context, email string) (err error) {
	s.logger.Info("Checking account existing in cache")
	repoAccount, err := s.registrationRepository.GetAccount(ctx, email)
	if err != nil {
		return
	}

	account := models.Account{
		Email:            email,
		Password:         repoAccount.Password,
		RegistrationDate: time.Now().In(time.UTC).In(time.UTC),
	}

	s.logger.Info("Creating account")
	tx, accountID, err := s.accountsRepository.CreateAccount(ctx, account)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return models.Error(models.Internal, err.Error())
	}

	err = s.accountEvents.AccountCreated(ctx, models.AccountCreatedDTO{
		ID:               accountID,
		Email:            account.Email,
		RegistrationDate: account.RegistrationDate,
		Username:         repoAccount.Username,
	})
	if err != nil {
		return models.Error(models.Internal, err.Error())
	}

	// The error is not critical, the data will still be deleted from the repository.
	if err = s.registrationRepository.DeleteAccount(ctx, email); err != nil {
		s.logger.Error("error while deleting account from registration repository: ", err.Error())
	}

	return nil
}

func (s *accountsService) SignIn(ctx context.Context, dto models.SignInDTO) (sessionID string, err error) {
	s.logger.Info("Getting account by email")
	account, err := s.accountsRepository.GetAccountByEmail(ctx, dto.Email)
	if err != nil {
		return
	}

	s.logger.Info("Password and hash comparison")
	if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(dto.Password)); err != nil {
		err = models.Error(models.InvalidArgument, "invalid login or password")
		return
	}

	s.logger.Info("Caching session")
	sessionID = uuid.NewString()
	err = s.sessionsRepository.SetSession(ctx, &models.Session{
		SessionID:    sessionID,
		AccountID:    account.ID,
		MachineID:    dto.MachineID,
		ClientIP:     dto.ClientIP,
		LastActivity: time.Now().In(time.UTC)}, s.cfg.SessionTTL)
	if err != nil {
		return "", err
	}

	return
}

func (s *accountsService) GetAccountID(ctx context.Context,
	sessionID, machineID string) (accountID string, err error) {
	s.logger.Info("Checking session")
	cached, err := s.checkAndUpdateSession(ctx, machineID, sessionID)
	if err != nil {
		return "", err
	}

	accountID = cached.AccountID
	return accountID, nil
}

func (s *accountsService) Logout(ctx context.Context,
	sessionID, machineID string) (err error) {
	s.logger.Info("Checking session")
	session, err := s.checkSession(ctx, machineID, sessionID)
	if err != nil {
		return
	}

	err = s.sessionsRepository.TerminateSessions(ctx, []string{sessionID}, session.AccountID)
	return
}

func (s *accountsService) RequestChangePasswordToken(ctx context.Context,
	email, callbackURL string) (err error) {
	exist, err := s.accountsRepository.IsAccountWithEmailExist(ctx, email)
	if err != nil {
		return err
	}
	if !exist {
		err = models.Error(models.NotFound, "account not found")
		return
	}

	token, err := jwt.GenerateToken(email, s.cfg.ChangePasswordTokenSecret,
		s.cfg.ChangePasswordTokenTTL)
	if err != nil {
		err = models.Error(models.Internal, err.Error())
		return
	}

	err = s.tokenDeliveryMQ.RequestChangePasswordTokenDelivery(ctx, email, token, callbackURL, s.cfg.ChangePasswordTokenTTL)
	return
}

func (s *accountsService) ChangePassword(ctx context.Context, token, newPassword string) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "accountsService.ChangePassword")
	defer span.Finish()

	s.logger.Info("Parsing jwt token")
	email, err := jwt.ParseToken(token, s.cfg.ChangePasswordTokenSecret)
	if err != nil {
		err = models.Error(models.InvalidArgument, err.Error())
		return
	}

	s.logger.Info("Checking account existing in DB")
	exist, err := s.accountsRepository.IsAccountWithEmailExist(ctx, email)
	if err != nil {
		return err
	}
	if !exist {
		err = models.Error(models.NotFound, "account not found")
		return
	}

	s.logger.Info("Generating hash for incoming password")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword),
		s.cfg.BcryptCost)
	if err != nil {
		err = models.Error(models.Internal, "can't generate password hash.")
		return
	}

	s.logger.Info("Changing account password")
	err = s.accountsRepository.ChangePassword(ctx, email, string(passwordHash))
	return
}

func (s *accountsService) GetAllSessions(ctx context.Context,
	sessionID, machineID string) (sessions map[string]*models.SessionInfo, err error) {
	s.logger.Info("Checking session")
	cache, err := s.checkAndUpdateSession(ctx, machineID, sessionID)
	if err != nil {
		return
	}

	sessions, err = s.sessionsRepository.GetSessionsForAccount(ctx, cache.AccountID)
	if err != nil {
		return
	}

	return
}

func (s *accountsService) TerminateSessions(ctx context.Context,
	sessionID, machineID string, sessionsToTerminateIds []string) (err error) {
	s.logger.Info("Checking session")
	session, err := s.checkAndUpdateSession(ctx, machineID, sessionID)
	if err != nil {
		return
	}

	var logout bool
	for i := range sessionsToTerminateIds {
		if sessionID == sessionsToTerminateIds[i] {
			logout = true
			break
		}
	}
	if !logout {
		go s.updateSession(context.Background(), &session, time.Now().In(time.UTC))
	}

	s.logger.Info("Terminating sessions")
	if err = s.sessionsRepository.TerminateSessions(ctx, sessionsToTerminateIds, session.AccountID); err != nil {
		return
	}

	return
}

func (s *accountsService) DeleteAccount(ctx context.Context, sessionID, machineID string) (err error) {
	session, err := s.checkSession(ctx, machineID, sessionID)
	if err != nil {
		return err
	}

	email, err := s.accountsRepository.GetAccountEmail(ctx, session.AccountID)
	if err != nil {
		return err
	}

	tx, err := s.accountsRepository.DeleteAccount(ctx, session.AccountID)
	if err != nil {
		return err
	}

	err = s.accountEvents.AccountDeleted(ctx, email, session.AccountID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return models.Error(models.Internal, err.Error())
	}

	go func(session models.Session) {
		for i := uint32(0); i < s.cfg.NumRetriesForTerminateSessions; i++ {
			terr := s.sessionsRepository.TerminateAllSessions(context.Background(), session.AccountID)
			if terr == nil || models.Code(terr) == models.NotFound {
				return
			}
			time.Sleep(s.cfg.RetrySleepTimeForTerminateSessions)
		}
	}(session)

	return nil
}

func (s *accountsService) checkSession(ctx context.Context, machineID, sessionID string) (session models.Session, err error) {
	s.logger.Info("Getting session cache")
	session, err = s.sessionsRepository.GetSession(ctx, sessionID)
	if err != nil {
		return
	}

	if machineID != session.MachineID {
		err = models.Error(models.Unauthenticated, "invalid session or machine id")
		session = models.Session{}
		return
	}
	return
}

func (s *accountsService) updateSession(ctx context.Context, session *models.Session, lastActivityTime time.Time) {
	s.logger.Info("Updating last activity for session")
	err := s.sessionsRepository.UpdateLastActivityForSession(ctx,
		session, lastActivityTime, s.cfg.SessionTTL)
	if err != nil && models.Code(err) != models.NotFound {
		s.logger.Warning("Session last activity not updated, error: ", err.Error())
	}
}

func (s *accountsService) checkAndUpdateSession(ctx context.Context, machineID, sessionID string) (session models.Session, err error) {
	s.logger.Info("Getting session cache")
	session, err = s.sessionsRepository.GetSession(ctx, sessionID)
	if err != nil {
		return
	}

	if machineID != session.MachineID {
		err = models.Error(models.Unauthenticated, "invalid session or machine id")
		session = models.Session{}
		return
	}

	go s.updateSession(context.Background(), &session, time.Now().In(time.UTC))
	return
}
