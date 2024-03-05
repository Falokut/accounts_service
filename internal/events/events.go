package events

import (
	"context"
	"errors"
	"time"

	"github.com/Falokut/accounts_service/internal/models"
)

type KafkaConfig struct {
	Brokers []string
}

func getContextError(ctx context.Context) (err error) {
	if ctx.Err() == nil {
		return nil
	}
	var code models.ErrorCode
	switch {
	case errors.Is(ctx.Err(), context.Canceled):
		code = models.Canceled
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		code = models.DeadlineExceeded
	}
	err = models.Error(code, ctx.Err().Error())
	return
}

type AccountsEventsMQ interface {
	AccountCreated(ctx context.Context, account models.AccountCreatedDTO) error
	AccountDeleted(ctx context.Context, email, accountID string) error
}

type TokensDeliveryMQ interface {
	RequestEmailVerificationTokenDelivery(ctx context.Context, email, token, callbackURL string, callbackURLTTL time.Duration) error
	RequestChangePasswordTokenDelivery(ctx context.Context, email, token, callbackURL string, callbackURLTTL time.Duration) error
}
