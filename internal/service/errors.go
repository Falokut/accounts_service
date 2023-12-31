package service

import (
	"errors"
	"fmt"

	accounts_service "github.com/Falokut/accounts_service/pkg/accounts_service/v1/protos"
	"github.com/Falokut/grpc_errors"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound                = errors.New("not found")
	ErrNoCtxMetaData           = errors.New("no context metadata provided")
	ErrInvalidSessionId        = errors.New("invalid session id")
	ErrAlreadyExist            = errors.New("already exist")
	ErrInvalidMachineID        = errors.New("invalid machine id")
	ErrInvalidClientIP         = errors.New("invalid client ip")
	ErrAccessDenied            = errors.New("access denied. Invalid session or machine id")
	ErrInternal                = errors.New("internal error")
	ErrAccountAlreadyActivated = errors.New("account already activated")
	ErrInvalidArgument         = errors.New("invalid input data")
	ErrFailedValidation        = errors.New("validation failed")
	ErrSessisonNotFound        = errors.New("session with specified id not found")
)

var errorCodes = map[error]codes.Code{
	redis.Nil:                  codes.NotFound,
	ErrNotFound:                codes.NotFound,
	ErrInvalidArgument:         codes.InvalidArgument,
	ErrNoCtxMetaData:           codes.Unauthenticated,
	ErrInvalidSessionId:        codes.Unauthenticated,
	ErrSessisonNotFound:        codes.Unauthenticated,
	ErrAlreadyExist:            codes.AlreadyExists,
	ErrInvalidMachineID:        codes.InvalidArgument,
	ErrInvalidClientIP:         codes.InvalidArgument,
	ErrFailedValidation:        codes.InvalidArgument,
	ErrAccessDenied:            codes.PermissionDenied,
	ErrInternal:                codes.Internal,
	ErrAccountAlreadyActivated: codes.AlreadyExists,
}

type errorHandler struct {
	logger *logrus.Logger
}

func newErrorHandler(logger *logrus.Logger) errorHandler {
	return errorHandler{
		logger: logger,
	}
}

func (e *errorHandler) createErrorResponceWithSpan(span opentracing.Span, err error, developerMessage string) error {
	if err == nil {
		return nil
	}

	span.SetTag("grpc.status", grpc_errors.GetGrpcCode(err))
	ext.LogError(span, err)
	return e.createErrorResponce(err, developerMessage)
}

func (e *errorHandler) createErrorResponce(err error, developerMessage string) error {
	var msg string
	if len(developerMessage) == 0 {
		msg = err.Error()
	} else {
		msg = fmt.Sprintf("%s. error: %v", developerMessage, err)
	}

	err = status.Error(grpc_errors.GetGrpcCode(err), msg)
	e.logger.Error(err)
	return err
}

func (e *errorHandler) createExtendedErrorResponceWithSpan(span opentracing.Span,
	err error, developerMessage, userMessage string) error {
	if err == nil {
		return nil
	}

	span.SetTag("grpc.status", grpc_errors.GetGrpcCode(err))
	ext.LogError(span, err)
	return e.createExtendedErrorResponce(err, developerMessage, userMessage)
}

func (e *errorHandler) createExtendedErrorResponce(err error, developerMessage, userMessage string) error {
	var msg string
	if developerMessage != "" {
		msg = fmt.Sprintf("%s. error: %v", developerMessage, err)
	} else {
		msg = err.Error()
	}

	extErr := status.New(grpc_errors.GetGrpcCode(err), msg)
	if len(userMessage) > 0 {
		extErr, _ = extErr.WithDetails(&accounts_service.UserErrorMessage{Message: userMessage})
		if extErr == nil {
			e.logger.Error(err)
			return err
		}
	}

	e.logger.Error(extErr)
	return extErr.Err()
}

func init() {
	grpc_errors.RegisterErrors(errorCodes)
}
