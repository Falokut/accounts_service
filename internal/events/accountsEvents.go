package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Falokut/accounts_service/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type accountsEvents struct {
	eventsWriter *kafka.Writer
	logger       *logrus.Logger
}

func NewAccountsEvents(cfg KafkaConfig, logger *logrus.Logger) *accountsEvents {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Logger:                 logger,
		AllowAutoTopicCreation: true,
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		Balancer:     &kafka.LeastBytes{},
	}
	return &accountsEvents{eventsWriter: w, logger: logger}
}

const (
	accountCreatedTopic = "account_created"
	accountDeletedTopic = "account_deleted"
)

func (e *accountsEvents) Shutdown() error {
	return e.eventsWriter.Close()
}

func (e *accountsEvents) AccountCreated(ctx context.Context, account models.AccountCreatedDTO) (err error) {
	defer e.handleError(ctx, &err)
	defer e.logError(err, "AccountCreated")

	body, err := json.Marshal(account)
	if err != nil {
		e.logger.Panic(err)
		return
	}

	err = e.eventsWriter.WriteMessages(ctx, kafka.Message{
		Topic: accountCreatedTopic,
		Key:   []byte(fmt.Sprint("account_", account.Id)),
		Value: body,
	})

	return
}

func (e *accountsEvents) AccountDeleted(ctx context.Context, email, accountId string) (err error) {
	defer e.handleError(ctx, &err)
	defer e.logError(err, "AccountDeleted")

	body, err := json.Marshal(struct {
		Email     string `json:"email"`
		AccountId string `json:"account_id"`
	}{
		Email:     email,
		AccountId: accountId,
	})
	if err != nil {
		e.logger.Panic(err)
		return
	}

	err = e.eventsWriter.WriteMessages(ctx, kafka.Message{
		Topic: accountDeletedTopic,
		Key:   []byte(fmt.Sprint("account_", accountId)),
		Value: body,
	})

	return
}

func (e *accountsEvents) handleError(ctx context.Context, err *error) {
	if ctx.Err() != nil {
		var code models.ErrorCode
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			code = models.Canceled
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			code = models.DeadlineExceeded
		}
		*err = models.Error(code, ctx.Err().Error())
		return
	}

	if err == nil || *err == nil {
		return
	}

	var serviceErr = &models.ServiceError{}
	if !errors.As(*err, &serviceErr) {
		*err = models.Error(models.Internal, "error while sending event notification")
	}
}

func (e *accountsEvents) logError(err error, functionName string) {
	if err == nil {
		return
	}

	var eventsErr = &models.ServiceError{}
	if errors.As(err, &eventsErr) {
		e.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           eventsErr.Msg,
				"error.code":          eventsErr.Code,
			},
		).Error("account events error occurred")
	} else {
		e.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("account events error occurred")
	}
}
