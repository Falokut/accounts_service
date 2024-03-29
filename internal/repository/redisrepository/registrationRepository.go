package redisrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Falokut/accounts_service/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RegistrationRepository struct {
	rdb     *redis.Client
	logger  *logrus.Logger
	metrics Metrics
}

func (r *RegistrationRepository) PingContext(ctx context.Context) error {
	if err := r.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error while pinging registration repository: %w", err)
	}

	return nil
}

// NewRedisRegistrationRepository initializes a new instance of RegistrationRepository with the provided options and logger.
func NewRedisRegistrationRepository(opt *redis.Options, logger *logrus.Logger, metrics Metrics) (*RegistrationRepository, error) {
	logger.Info("Creating registration repository client")
	rdb, err := NewRedisClient(opt)
	if err != nil {
		return nil, err
	}
	return &RegistrationRepository{
		rdb:     rdb,
		logger:  logger,
		metrics: metrics,
	}, nil
}

// Shutdown gracefully shuts down the registration repository.
func (r *RegistrationRepository) Shutdown() {
	r.logger.Info("Registration repository shutting down")
	err := r.rdb.Close()
	if err != nil {
		r.logger.Errorf("error while shutting down registration repository %v", err)
	}
}

// IsAccountExist checks if the provided email account is present in the repository.
func (r *RegistrationRepository) IsAccountExist(ctx context.Context, email string) (inCache bool, err error) {
	defer r.updateMetrics(err, "IsAccountExist")
	defer handleError(ctx, &err)
	defer r.logError(err, "IsAccountExist")
	num, err := r.rdb.Exists(ctx, email).Result()
	if err != nil {
		return
	}

	return num > 0, nil
}

// SetAccount caches the provided account information with the specified email in the repository.
// It marshals the account data into JSON and sets it in the repository with the specified TTL.
func (r *RegistrationRepository) SetAccount(ctx context.Context,
	email string, account models.RegisteredAccount, ttl time.Duration) (err error) {
	defer r.updateMetrics(err, "SetAccount")
	defer handleError(ctx, &err)
	defer r.logError(err, "SetAccount")

	r.logger.Info("Marshaling data")
	serialized, err := json.Marshal(&account)
	if err != nil {
		return
	}

	_, err = r.rdb.Set(ctx, email, serialized, ttl).Result()
	return nil
}

// GetAccount retrieves the cached account information for the specified email from the repository.
// It returns the cached account data and any encountered error during retrieval.
func (r *RegistrationRepository) GetAccount(ctx context.Context, email string) (account models.RegisteredAccount, err error) {
	defer r.updateMetrics(err, "GetAccount")
	defer handleError(ctx, &err)
	defer r.logError(err, "GetAccount")

	body, err := r.rdb.Get(ctx, email).Bytes()
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &account)
	return
}

// DeleteAccount deletes the account information associated with the specified email from the repository.
func (r *RegistrationRepository) DeleteAccount(ctx context.Context, email string) (err error) {
	defer r.updateMetrics(err, "DeleteAccount")
	defer handleError(ctx, &err)
	defer r.logError(err, "DeleteAccount")

	err = r.rdb.Del(ctx, email).Err()
	return
}

func (r *RegistrationRepository) logError(err error, functionName string) {
	if err == nil {
		return
	}

	var repoErr = &models.ServiceError{}
	if errors.As(err, &repoErr) {
		r.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           repoErr.Msg,
				"error.code":          repoErr.Code,
			},
		).Error("registration repository error occurred")
	} else {
		r.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("registration repository error occurred")
	}
}

func (r *RegistrationRepository) updateMetrics(err error, functionName string) {
	if err == nil {
		r.metrics.IncCacheHits(functionName)
		return
	}
	if models.Code(err) == models.NotFound {
		r.metrics.IncCacheMiss(functionName)
	}
}
