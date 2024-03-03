package events

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Falokut/accounts_service/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type tokensDeliveryMQ struct {
	eventsWriter *kafka.Writer
	logger       *logrus.Logger
}

func NewTokensDeliveryMQ(cfg KafkaConfig, logger *logrus.Logger) *tokensDeliveryMQ {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Logger:                 logger,
		AllowAutoTopicCreation: true,
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		Balancer:     &kafka.LeastBytes{},
	}
	return &tokensDeliveryMQ{eventsWriter: w, logger: logger}
}

const (
	emailVerificationTopic = "email_verification_delivery_request"
	passwordChangeTopic    = "password_change_delivery_request"
)

type tokenDeviveryRequest struct {
	Email          string        `json:"email"`
	Token          string        `json:"token"`
	CallbackUrl    string        `json:"callback_url"`
	CallbackUrlTtl time.Duration `json:"callback_url_ttl"`
}

func (e *tokensDeliveryMQ) RequestEmailVerificationTokenDelivery(ctx context.Context,
	email, token, callbackUrl string, callbackUrlTtl time.Duration) (err error) {
	defer e.handleError(ctx, &err)
	defer e.logError(err, "RequestEmailVerificationTokenDelivery")

	body, err := json.Marshal(tokenDeviveryRequest{
		Email:          email,
		Token:          token,
		CallbackUrl:    callbackUrl,
		CallbackUrlTtl: callbackUrlTtl,
	})
	if err != nil {
		e.logger.Panic(err)
	}

	err = e.eventsWriter.WriteMessages(ctx, kafka.Message{
		Topic: emailVerificationTopic,
		Key:   []byte(email),
		Value: body,
	})

	return
}

func (e *tokensDeliveryMQ) RequestChangePasswordTokenDelivery(ctx context.Context,
	email, token, callbackUrl string, callbackUrlTtl time.Duration) (err error) {
	defer e.handleError(ctx, &err)
	defer e.logError(err, "RequestChangePasswordTokenDelivery")

	body, err := json.Marshal(tokenDeviveryRequest{
		Email:          email,
		Token:          token,
		CallbackUrl:    callbackUrl,
		CallbackUrlTtl: callbackUrlTtl,
	})
	if err != nil {
		e.logger.Panic(err)
		return
	}

	err = e.eventsWriter.WriteMessages(ctx, kafka.Message{
		Topic: passwordChangeTopic,
		Key:   []byte(email),
		Value: body,
	})

	return
}

func (e *tokensDeliveryMQ) Shutdown() error {
	return e.eventsWriter.Close()
}

func (e *tokensDeliveryMQ) logError(err error, functionName string) {
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
		).Error("tokens delivery error occurred")
	} else {
		e.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("tokens delivery error occurred")
	}
}

func (e *tokensDeliveryMQ) handleError(ctx context.Context, err *error) {
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
		*err = models.Error(models.Internal, "error while sending message in MQ")
	}
}
