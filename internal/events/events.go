package events

import (
	"context"
	"time"

	"github.com/Falokut/accounts_service/internal/models"
)

type KafkaConfig struct {
	Brokers []string
}

type AccountsEventsMQ interface {
	AccountCreated(ctx context.Context, account models.Account) error
	AccountDeleted(ctx context.Context, email, accountId string) error
}

type TokensDeliveryMQ interface {
	RequestEmailVerificationTokenDelivery(ctx context.Context, email, token, callbackUrl string, callbackUrlTtl time.Duration) error
	RequestChangePasswordTokenDelivery(ctx context.Context, email, token, callbackUrl string, callbackUrlTtl time.Duration) error
}
