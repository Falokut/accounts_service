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
	"golang.org/x/exp/maps"
)

type SessionsRepository struct {
	rdb     *redis.Client
	logger  *logrus.Logger
	metrics Metrics
}

// NewSessionsRepository creates a new session database using the provided Redis options, logger, and session TTL.
// It initializes two Redis clients for session and account session caching, and verifies the connection to each Redis instance.
func NewSessionsRepository(opt *redis.Options,
	logger *logrus.Logger,
	metrics Metrics,
) (*SessionsRepository, error) {
	logger.Info("Creating session repository client")

	rdb := redis.NewClient(opt)
	if rdb == nil {
		return nil, errors.New("can't create new redis client")
	}

	logger.Info("Pinging sessions database client")
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("connection is not established: %s", err.Error())
	}

	return &SessionsRepository{
		rdb:     rdb,
		logger:  logger,
		metrics: metrics,
	}, nil
}

func getKeyForAccountSessionsList(accountId string) string {
	return "account_" + accountId
}

func (r *SessionsRepository) removeNonexistantKeys(ctx context.Context,
	accountId string, keys []string) (existsKeys []string, err error) {
	defer r.updateMetrics(err, "removeNonexistantKeys")
	defer handleError(ctx, &err)
	defer r.logError(err, "removeNonexistantKeys")

	if len(keys) == 0 {
		return []string{}, nil
	}

	res, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return
	}

	var exists = make(map[string]struct{}, len(keys))
	for i := range res {
		if res[i] == nil {
			continue
		}
		exists[res[i].(string)] = struct{}{}
	}

	toRemoveLen := len(keys) - len(exists)
	if toRemoveLen == 0 {
		return keys, nil
	}

	var toRemove = make([]string, toRemoveLen)
	for i := range keys {
		if _, ok := exists[keys[i]]; !ok {
			toRemove = append(toRemove, keys[i])
		}
	}

	err = r.rdb.SRem(ctx, getKeyForAccountSessionsList(accountId), toRemove).Err()
	if err != nil {
		return
	}

	existsKeys = maps.Keys(exists)
	return
}

func (r *SessionsRepository) PingContext(ctx context.Context) error {
	if err := r.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error while pinging sessions database: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the sessions database repository by closing the Redis client for session caching.
func (r *SessionsRepository) Shutdown() error {
	r.logger.Info("Sessions database repository shutting down")
	return r.rdb.Close()
}

func (r *SessionsRepository) SetSession(ctx context.Context, session models.Session, ttl time.Duration) (err error) {
	defer r.updateMetrics(err, "SetSession")
	defer handleError(ctx, &err)
	defer r.logError(err, "SetSession")

	r.logger.Info("Marshalling data")
	serialized, err := json.Marshal(session)
	if err != nil {
		return
	}

	tx := r.rdb.Pipeline()
	r.logger.Info("Caching sessions data")
	err = tx.Set(ctx, session.SessionId, serialized, ttl).Err()
	if err != nil {
		return
	}

	r.logger.Info("Adding session to accounts sessions list")
	listKey := getKeyForAccountSessionsList(session.AccountId)
	err = tx.SAdd(ctx, listKey, session.SessionId).Err()
	if err != nil {
		return
	}
	err = tx.Expire(ctx, listKey, ttl).Err()
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx)
	return
}

func (r *SessionsRepository) TerminateSessions(ctx context.Context, sessionsIds []string, accountId string) (err error) {
	defer r.updateMetrics(err, "TerminateSessions")
	defer handleError(ctx, &err)
	defer r.logError(err, "TerminateSessions")

	accountSessions, err := r.GetSessionsForAccount(ctx, accountId)
	if err != nil {
		return
	}

	toDelete := make([]string, 0, len(sessionsIds))
	for i := range sessionsIds {
		_, ok := accountSessions[sessionsIds[i]]
		if ok {
			toDelete = append(toDelete, sessionsIds[i])
		}
	}
	if len(toDelete) == 0 {
		err = models.Error(models.InvalidArgument, "invalid sessions ids, not found any")
		return
	}

	tx := r.rdb.Pipeline()
	err = tx.Del(ctx, toDelete...).Err()
	if err != nil {
		return
	}

	for _, sessionId := range toDelete {
		err = tx.SRem(ctx, sessionId).Err()
		if err != nil {
			return
		}
	}
	_, err = tx.Exec(ctx)
	return
}

func (r *SessionsRepository) GetSessionsIds(ctx context.Context, accountId string) (sessionsIds []string, err error) {
	defer r.updateMetrics(err, "GetSessionsIds")
	defer handleError(ctx, &err)
	defer r.logError(err, "GetSessionsIds")

	sessionsIds, err = r.rdb.SMembers(ctx, getKeyForAccountSessionsList(accountId)).Result()
	if err != nil || len(sessionsIds) == 0 {
		return
	}

	sessionsIds, err = r.removeNonexistantKeys(ctx, accountId, sessionsIds)
	return
}

// GetSession retrieves the session database for a given session ID from the Redis database.
// It retrieves the session data from the database, and unmarshals it into a models.Session.
// If the session data is not found in the database, it returns an error indicating that the session was not found.
func (r *SessionsRepository) GetSession(ctx context.Context, sessionId string) (session models.Session, err error) {
	defer r.updateMetrics(err, "GetSession")
	defer handleError(ctx, &err)
	defer r.logError(err, "GetSession")

	body, err := r.rdb.Get(ctx, sessionId).Bytes()
	if err != nil {
		return
	}

	r.logger.Info("Unmarshal database data")
	if err = json.Unmarshal(body, &session); err != nil {
		return
	}

	return
}

// UpdateLastActivityForSession updates the last activity time for a cached session in the Redis database.
// It updates the LastActivity field of the cached session, and then caches the updated session.
func (r *SessionsRepository) UpdateLastActivityForSession(ctx context.Context,
	cachedSession models.Session, lastActivityTime time.Time, ttl time.Duration) (err error) {
	defer r.updateMetrics(err, "UpdateLastActivityForSession")
	defer handleError(ctx, &err)
	defer r.logError(err, "UpdateLastActivityForSession")

	cachedSession.LastActivity = lastActivityTime
	err = r.SetSession(ctx, cachedSession, ttl)
	return
}

// GetSessionsForAccount retrieves the sessions associated with the specified account from the Redis database.
func (r *SessionsRepository) GetSessionsForAccount(ctx context.Context,
	accountId string) (sessions map[string]models.SessionInfo, err error) {
	defer r.updateMetrics(err, "GetSessionsForAccount")
	defer handleError(ctx, &err)
	defer r.logError(err, "GetSessionsForAccount")

	sessionsIds, err := r.rdb.SMembers(ctx, getKeyForAccountSessionsList(accountId)).Result()
	if err != nil {
		return
	}

	sessionsInfo, err := r.rdb.MGet(ctx, sessionsIds...).Result()
	if err != nil {
		return
	}

	sessions = make(map[string]models.SessionInfo, len(sessionsInfo))
	for i := range sessionsInfo {
		var database models.Session
		err = json.Unmarshal([]byte(sessionsInfo[i].(string)), &database)
		if err != nil {
			return
		}

		sessions[database.SessionId] = models.SessionInfo{
			ClientIp:     database.ClientIp,
			MachineId:    database.MachineId,
			LastActivity: database.LastActivity,
		}
	}

	return
}

type AccountSessions struct {
	Sessions []string `json:"sessions"`
}

func (r *SessionsRepository) TerminateAllSessions(ctx context.Context, accountId string) (err error) {
	defer r.updateMetrics(err, "TerminateAllSessions")
	defer handleError(ctx, &err)
	defer r.logError(err, "TerminateAllSessions")

	listKey := getKeyForAccountSessionsList(accountId)
	sessionsIds, err := r.rdb.SMembers(ctx, listKey).Result()
	if errors.Is(err, redis.Nil) || len(sessionsIds) == 0 {
		// Not an error, just an empty sessions map for the account
		return nil
	}

	tx := r.rdb.Pipeline()
	err = tx.Del(ctx, sessionsIds...).Err()
	if err != nil {
		return
	}
	err = tx.Del(ctx, listKey).Err()
	if err != nil {
		return
	}
	_, err = tx.Exec(ctx)
	return
}

func (r *SessionsRepository) logError(err error, functionName string) {
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
		).Error("sessions database error occurred")
	} else {
		r.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("sessions database error occurred")
	}
}

func (r *SessionsRepository) updateMetrics(err error, functionName string) {
	if err == nil {
		r.metrics.IncCacheHits(functionName)
		return
	}
	if models.Code(err) == models.NotFound {
		r.metrics.IncCacheMiss(functionName)
	}
}
