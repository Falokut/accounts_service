package redisrepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Falokut/accounts_service/internal/models"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(opt *redis.Options) (*redis.Client, error) {
	rdb := redis.NewClient(opt)
	if rdb == nil {
		return nil, errors.New("can't create new redis client")
	}

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("connection is not established: %s", err.Error())
	}

	return rdb, nil
}

func handleError(ctx context.Context, err *error) {
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

	var repoErr = &models.ServiceError{}
	if !errors.As(*err, &repoErr) {
		var code models.ErrorCode
		switch {
		case errors.Is(*err, redis.Nil):
			code = models.NotFound
			*err = models.Error(code, "cache entity not found")
		default:
			code = models.Internal
			*err = models.Error(code, "cache internal error")
		}
	}
}

type Metrics interface {
	IncCacheHits(method string)
	IncCacheMiss(method string)
}
