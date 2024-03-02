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
	DeleteAccount(ctx context.Context, sessionId, machineId string) error
	RequestAccountVerificationToken(ctx context.Context, email, callbackUrl string) error
	VerifyAccount(ctx context.Context, token string) error
	SignIn(ctx context.Context, dto models.SignInDTO) (sessionId string, err error)
	GetAccountId(ctx context.Context, sessionId, machineId string) (accountId string, err error)
	Logout(ctx context.Context, sessionId, machineId string) error
	RequestChangePasswordToken(ctx context.Context, email, callbackUrl string) error
	ChangePassword(ctx context.Context, token, newPassword string) error
	GetAllSessions(ctx context.Context, sessionId, machineId string) (map[string]models.SessionInfo, error)
	TerminateSessions(ctx context.Context, sessionId, machineId string, sessionsToTerminateIds []string) error
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
	cfg AccountsServiceConfig) *accountsService {

	return &accountsService{accountsRepository: repo,
		logger:                 logger,
		registrationRepository: registrationRepository,
		sessionsRepository:     sessionsRepository,
		cfg:                    cfg,
		tokenDeliveryMQ:        tokenDeliveryMQ,
		accountEvents:          accountEvents,
	}
}

func (s *accountsService) CreateAccount(ctx context.Context,
	dto models.CreateAccountDTO) (err error) {

	exist, err := s.accountsRepository.IsAccountWithEmailExist(ctx, dto.Email)
	if err != nil {
		return
	}
	if exist {
		err = models.Error(models.Conflict, "a user with this email address already exists. "+
			"please try another one or simple log in")
		return
	}

	inCache, err := s.registrationRepository.IsAccountExist(ctx, dto.Email)
	if err != nil {
		return
	}
	if inCache {
		err = models.Error(models.Conflict, "a user with this email address already exists. "+
			"please try another one or verify email and log in")
		return
	}

	s.logger.Info("Generating hash from password")
	password_hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), s.cfg.BcryptCost)
	if err != nil {
		err = models.Error(models.Internal, "can't generate password hash")
		return
	}

	err = s.registrationRepository.SetAccount(ctx, dto.Email,
		models.RegisteredAccount{
			Username: dto.Username,
			Password: string(password_hash),
		},
		s.cfg.NonActivatedAccountTTL)

	if err != nil {
		return
	}

	return
}

func (s *accountsService) RequestAccountVerificationToken(ctx context.Context,
	email, callbackUrl string) (err error) {

	exist, err := s.accountsRepository.IsAccountWithEmailExist(ctx, email)
	if err != nil {
		return
	}
	if exist {
		err = models.Error(models.InvalidArgument, "account already activated")
		return
	}

	inCache, err := s.registrationRepository.IsAccountExist(ctx, email)
	if err != nil {
		return
	}
	if !inCache {
		err = models.Error(models.NotFound, "a account with this email address not exist")
		return
	}

	token, err := jwt.GenerateToken(email, s.cfg.VerifyAccountTokenSecret, s.cfg.VerifyAccountTokenTTL)
	if err != nil {
		return
	}

	err = s.tokenDeliveryMQ.RequestEmailVerificationTokenDelivery(ctx, email, token, callbackUrl, s.cfg.VerifyAccountTokenTTL)
	if err != nil {
		return
	}
	return
}

func (s *accountsService) VerifyAccount(ctx context.Context, token string) (err error) {
	s.logger.Info("Parsing token")
	email, err := jwt.ParseToken(token, config.GetConfig().JWT.VerifyAccountToken.Secret)
	if err != nil {
		err = models.Error(models.InvalidArgument, err.Error())
		return
	}

	err = s.createAccount(ctx, email)
	return
}

func (s *accountsService) createAccount(ctx context.Context, email string) (err error) {
	s.logger.Info("Checking account existing in cache")
	repoAccount, err := s.registrationRepository.GetAccount(ctx, email)
	if err != nil {
		return
	}

	account := models.Account{
		Email:            email,
		Password:         string(repoAccount.Password),
		RegistrationDate: time.Now().In(time.UTC).In(time.UTC),
	}

	s.logger.Info("Creating account")
	tx, accountId, err := s.accountsRepository.CreateAccount(ctx, account)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account.Password = ""
	account.Id = accountId

	err = s.accountEvents.AccountCreated(ctx, account)
	if err != nil {
		err = models.Error(models.Internal, err.Error())
		return
	}
	err = tx.Commit()
	if err != nil {
		err = models.Error(models.Internal, err.Error())
		return
	}

	// The error is not critical, the data will still be deleted from the repository.
	if err = s.registrationRepository.DeleteAccount(ctx, email); err != nil {
		s.logger.Warning("can't delete account from registration repository: ", err.Error())
	}

	return
}

func (s *accountsService) SignIn(ctx context.Context, dto models.SignInDTO) (sessionId string, err error) {
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
	sessionId = uuid.NewString()
	err = s.sessionsRepository.SetSession(ctx, models.Session{
		SessionId:    sessionId,
		AccountId:    account.Id,
		MachineId:    dto.MachineId,
		ClientIp:     dto.ClientIp,
		LastActivity: time.Now().In(time.UTC)}, s.cfg.SessionTTL)
	if err != nil {
		return "", err
	}

	return
}

func (s *accountsService) GetAccountId(ctx context.Context,
	sessionId, machineId string) (accountId string, err error) {

	s.logger.Info("Checking session")
	cached, err := s.checkSession(ctx, machineId, sessionId)
	if err != nil {
		return
	}

	accountId = cached.AccountId
	return
}

func (s *accountsService) Logout(ctx context.Context,
	sessionId, machineId string) (err error) {

	s.logger.Info("Checking session")
	session, err := s.checkSession(ctx, machineId, sessionId)
	if err != nil {
		return
	}

	err = s.sessionsRepository.TerminateSessions(ctx, []string{sessionId}, session.AccountId)
	return
}

func (s *accountsService) RequestChangePasswordToken(ctx context.Context,
	email, callbackUrl string) (err error) {

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

	err = s.tokenDeliveryMQ.RequestChangePasswordTokenDelivery(ctx, email, token, callbackUrl, s.cfg.ChangePasswordTokenTTL)
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
	password_hash, err := bcrypt.GenerateFromPassword([]byte(newPassword),
		s.cfg.BcryptCost)
	if err != nil {
		err = models.Error(models.Internal, "can't generate password hash.")
		return
	}

	s.logger.Info("Changing account password")
	err = s.accountsRepository.ChangePassword(ctx, email, string(password_hash))
	return
}

func (s *accountsService) GetAllSessions(ctx context.Context,
	sessionId, machineId string) (sessions map[string]models.SessionInfo, err error) {

	s.logger.Info("Checking session")
	cache, err := s.checkSession(ctx, machineId, sessionId)
	if err != nil {
		return
	}

	sessions, err = s.sessionsRepository.GetSessionsForAccount(ctx, cache.AccountId)
	if err != nil {
		return
	}

	return
}

func (s *accountsService) TerminateSessions(ctx context.Context,
	sessionId, machineId string, sessionsToTerminateIds []string) (err error) {

	s.logger.Info("Checking session")
	cache, err := s.checkSession(ctx, machineId, sessionId)
	if err != nil {
		return
	}

	s.logger.Info("Terminating sessions")
	if err = s.sessionsRepository.TerminateSessions(ctx, sessionsToTerminateIds, cache.AccountId); err != nil {
		return
	}

	return
}

func (s *accountsService) DeleteAccount(ctx context.Context, sessionId, machineId string) (err error) {
	cache, err := s.checkSession(ctx, machineId, sessionId)
	if err != nil {
		return
	}

	email, err := s.accountsRepository.GetAccountEmail(ctx, cache.AccountId)
	if err != nil {
		return
	}
	tx, err := s.accountsRepository.DeleteAccount(ctx, cache.AccountId)
	if err != nil {
		return
	}

	err = s.accountEvents.AccountDeleted(ctx, email, cache.AccountId)
	if err != nil {
		return
	}
	tx.Commit()

	go func() {
		for i := uint32(0); i < s.cfg.NumRetriesForTerminateSessions; i++ {
			err = s.sessionsRepository.TerminateAllSessions(context.Background(), cache.AccountId)
			if err == nil {
				return
			}
			time.Sleep(s.cfg.RetrySleepTimeForTerminateSessions)
		}
	}()

	return

}

// Also updates last activity time at background
func (s *accountsService) checkSession(ctx context.Context, machineId, sessionId string) (session models.Session, err error) {
	s.logger.Info("Getting session cache")
	session, err = s.sessionsRepository.GetSession(ctx, sessionId)
	if err != nil {
		return
	}

	if machineId != session.MachineId {
		err = models.Error(models.Unauthenticated, "invalid session or machine id")
		session = models.Session{}
		return
	}

	go func(session models.Session, lastActivityTime time.Time) {
		s.logger.Info("Updating last activity for session")
		err = s.sessionsRepository.UpdateLastActivityForSession(context.Background(),
			session, lastActivityTime, s.cfg.SessionTTL)
		if err != nil {
			s.logger.Warning("Session last activity not updated, error: ", err.Error())
		}
	}(session, time.Now().In(time.UTC))

	return
}
