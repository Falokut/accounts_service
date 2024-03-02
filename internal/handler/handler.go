package handler

import (
	"context"
	"errors"
	"net"

	"github.com/Falokut/accounts_service/internal/models"
	"github.com/Falokut/accounts_service/internal/service"
	accounts_service "github.com/Falokut/accounts_service/pkg/accounts_service/v1/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AccountsServiceHandler struct {
	accounts_service.UnimplementedAccountsServiceV1Server
	logger          *logrus.Logger
	accountsService service.AccountsService
}

func NewAccountsServiceHandler(logger *logrus.Logger,
	accountsService service.AccountsService) *AccountsServiceHandler {
	return &AccountsServiceHandler{
		logger:          logger,
		accountsService: accountsService,
	}
}

func (h *AccountsServiceHandler) CreateAccount(ctx context.Context,
	in *accounts_service.CreateAccountRequest) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	if err = validateSignupInput(in); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.accountsService.CreateAccount(ctx, models.CreateAccountDTO{
		Email:    in.Email,
		Username: in.Username,
		Password: in.Password,
	})
	if err != nil {
		return
	}

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) RequestAccountVerificationToken(ctx context.Context,
	in *accounts_service.VerificationTokenRequest) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	if err = validateEmail(in.Email); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.accountsService.RequestAccountVerificationToken(ctx, in.Email, in.URL)
	if err != nil {
		return
	}
	
	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) VerifyAccount(ctx context.Context,
	in *accounts_service.VerifyAccountRequest) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	err = h.accountsService.VerifyAccount(ctx, in.VerificationToken)
	if err != nil {
		return
	}
	return &emptypb.Empty{}, err
}

func (h *AccountsServiceHandler) SignIn(ctx context.Context,
	in *accounts_service.SignInRequest) (res *accounts_service.AccessResponse, err error) {
	defer h.handleError(&err)

	if net.ParseIP(in.ClientIp) == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client ip address")
	}

	machineID, err := h.getMachineIdFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	sessionId, err := h.accountsService.SignIn(ctx, models.SignInDTO{
		Email:     in.Email,
		Password:  in.Password,
		ClientIp:  in.ClientIp,
		MachineId: machineID,
	})

	return &accounts_service.AccessResponse{SessionID: sessionId}, nil
}

func (h *AccountsServiceHandler) GetAccountId(ctx context.Context,
	in *emptypb.Empty) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	sessionId, machineId, err := h.getAuthHeaders(ctx)
	if err != nil {
		return
	}

	accountId, err := h.accountsService.GetAccountId(ctx, sessionId, machineId)
	if err != nil {
		return
	}

	header := metadata.Pairs(AccountIdContext, accountId)
	grpc.SetHeader(ctx, header)

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) Logout(ctx context.Context,
	in *emptypb.Empty) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	sessionId, machineId, err := h.getAuthHeaders(ctx)
	if err != nil {
		return
	}

	err = h.accountsService.Logout(ctx, sessionId, machineId)
	if err != nil {
		return
	}

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) RequestChangePasswordToken(ctx context.Context,
	in *accounts_service.ChangePasswordTokenRequest) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	err = h.accountsService.RequestChangePasswordToken(ctx, in.Email, in.URL)
	if err != nil {
		return
	}

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) ChangePassword(ctx context.Context,
	in *accounts_service.ChangePasswordRequest) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	err = validatePassword(in.NewPassword)
	if err != nil {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	err = h.accountsService.ChangePassword(ctx, in.ChangePasswordToken, in.NewPassword)
	if err != nil {
		return
	}

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) GetAllSessions(ctx context.Context,
	in *emptypb.Empty) (res *accounts_service.AllSessionsResponse, err error) {
	defer h.handleError(&err)

	sessionId, machineId, err := h.getAuthHeaders(ctx)
	if err != nil {
		return
	}

	sessions, err := h.accountsService.GetAllSessions(ctx, sessionId, machineId)
	sessionsInfo := make(map[string]*accounts_service.SessionInfo, len(sessions))
	for key, info := range sessions {
		sessionsInfo[key] = &accounts_service.SessionInfo{
			ClientIp:     info.ClientIp,
			MachineId:    info.MachineId,
			LastActivity: timestamppb.New(info.LastActivity.UTC()),
		}
	}

	return &accounts_service.AllSessionsResponse{Sessions: sessionsInfo}, nil
}

func (h *AccountsServiceHandler) TerminateSessions(ctx context.Context,
	in *accounts_service.TerminateSessionsRequest) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	sessionId, machineId, err := h.getAuthHeaders(ctx)
	if err != nil {
		return
	}
	err = h.accountsService.TerminateSessions(ctx, sessionId, machineId, in.SessionsToTerminate)
	if err != nil {
		return
	}

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) DeleteAccount(ctx context.Context,
	in *emptypb.Empty) (res *emptypb.Empty, err error) {
	defer h.handleError(&err)

	sessionId, machineId, err := h.getAuthHeaders(ctx)
	if err != nil {
		return
	}

	err = h.accountsService.DeleteAccount(ctx, sessionId, machineId)
	if err != nil {
		return
	}

	return &emptypb.Empty{}, nil
}

func (h *AccountsServiceHandler) getAuthHeaders(ctx context.Context) (sessionId, machineId string, err error) {
	sessionId, err = h.getSessionIdFromCtx(ctx)
	if err != nil {
		return
	}

	h.logger.Info("Getting client ip from ctx")
	machineId, err = h.getMachineIdFromCtx(ctx)
	if err != nil {
		return
	}

	return
}

// --------------------- CONTEXTS ---------------------
const (
	AccountIdContext = "X-Account-Id"
	SessionIdContext = "X-Session-Id"
	MachineIdContext = "X-Machine-Id"
)

//-----------------------------------------------------

func (h *AccountsServiceHandler) getSessionIdFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no context metadata provided")
	}

	sessionID := md.Get(SessionIdContext)
	if len(sessionID) == 0 || sessionID[0] == "" {
		return "", status.Error(codes.Unauthenticated, "no session id provided")
	}

	return sessionID[0], nil
}

func (h *AccountsServiceHandler) getMachineIdFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no context metadata provided")
	}

	machineID := md.Get(MachineIdContext)
	if len(machineID) == 0 || machineID[0] == "" {
		return "", status.Error(codes.Unauthenticated, "no machine id provided")
	}

	return machineID[0], nil
}

func (h *AccountsServiceHandler) handleError(err *error) {
	if err == nil || *err == nil {
		return
	}

	serviceErr := &models.ServiceError{}
	if errors.As(*err, &serviceErr) {
		*err = status.Error(convertServiceErrCodeToGrpc(serviceErr.Code), serviceErr.Msg)
	} else if _, ok := status.FromError(*err); !ok {
		e := *err
		*err = status.Error(codes.Unknown, e.Error())
	}
}

func convertServiceErrCodeToGrpc(code models.ErrorCode) codes.Code {
	switch code {
	case models.Internal:
		return codes.Internal
	case models.InvalidArgument:
		return codes.InvalidArgument
	case models.Unauthenticated:
		return codes.Unauthenticated
	case models.Conflict:
		return codes.AlreadyExists
	case models.NotFound:
		return codes.NotFound
	case models.Canceled:
		return codes.Canceled
	case models.DeadlineExceeded:
		return codes.DeadlineExceeded
	case models.PermissionDenied:
		return codes.PermissionDenied
	default:
		return codes.Unknown
	}
}
