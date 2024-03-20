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

// NewSessionsRepository creates a new session repository using the provided Redis options, logger, and session TTL.
// It initializes two Redis clients for session and account session caching, and verifies the connection to each Redis instance.
func NewSessionsRepository(opt *redis.Options,
	logger *logrus.Logger,
	metrics Metrics,
) (*SessionsRepository, error) {
	logger.Info("Creating session repository client")

	rdb, err := NewRedisClient(opt)
	if err != nil {
		return nil, err
	}

	return &SessionsRepository{
		rdb:     rdb,
		logger:  logger,
		metrics: metrics,
	}, nil
}

func getKeyForAccountSessionsList(accountID string) string {
	return "account_" + accountID
}

func (r *SessionsRepository) removeNonexistantKeys(ctx context.Context,
	accountID string, keys []string) (existsKeys []string, err error) {
	defer r.updateMetrics(&err, "removeNonexistantKeys")
	defer handleError(ctx, &err)
	defer r.logError(&err, "removeNonexistantKeys")

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

	err = r.rdb.SRem(ctx, getKeyForAccountSessionsList(accountID), toRemove).Err()
	if err != nil {
		return
	}

	existsKeys = maps.Keys(exists)
	return
}

func (r *SessionsRepository) PingContext(ctx context.Context) error {
	if err := r.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error while pinging sessions repository: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the sessions repository by closing the Redis client for session caching.
func (r *SessionsRepository) Shutdown() {
	r.logger.Info("Sessions repository shutting down")
	err := r.rdb.Close()
	if err != nil {
		r.logger.Errorf("error while shutting down sessions repository %v", err)
	}
}

func (r *SessionsRepository) SetSession(ctx context.Context, session *models.Session, ttl time.Duration) (err error) {
	defer r.updateMetrics(&err, "SetSession")
	defer handleError(ctx, &err)
	defer r.logError(&err, "SetSession")

	r.logger.Info("Marshaling data")
	serialized, err := json.Marshal(*session)
	if err != nil {
		return
	}

	tx := r.rdb.Pipeline()
	r.logger.Info("Caching sessions data")
	err = tx.Set(ctx, session.SessionID, serialized, ttl).Err()
	if err != nil {
		return
	}

	r.logger.Info("Adding session to accounts sessions list")
	listKey := getKeyForAccountSessionsList(session.AccountID)
	err = tx.SAdd(ctx, listKey, session.SessionID).Err()
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

func (r *SessionsRepository) TerminateSessions(ctx context.Context, sessionsIds []string, accountID string) (err error) {
	defer r.updateMetrics(&err, "TerminateSessions")
	defer handleError(ctx, &err)
	defer r.logError(&err, "TerminateSessions")

	if len(sessionsIds) == 0 {
		err = models.Error(models.InvalidArgument, "invalid sessions ids, it mustn't be empty")
		return
	}

	sessionsIds, err = r.removeNonexistantKeys(ctx, accountID, sessionsIds)
	if len(sessionsIds) == 0 {
		return nil
	}

	tx := r.rdb.Pipeline()
	tx.Del(ctx, sessionsIds...)
	err = tx.SRem(ctx, getKeyForAccountSessionsList(accountID), sessionsIds).Err()
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx)
	return
}

func (r *SessionsRepository) GetSessionsIds(ctx context.Context, accountID string) (sessionsIds []string, err error) {
	defer r.updateMetrics(&err, "GetSessionsIds")
	defer handleError(ctx, &err)
	defer r.logError(&err, "GetSessionsIds")

	sessionsIds, err = r.rdb.SMembers(ctx, getKeyForAccountSessionsList(accountID)).Result()
	if err != nil || len(sessionsIds) == 0 {
		return
	}

	sessionsIds, err = r.removeNonexistantKeys(ctx, accountID, sessionsIds)
	return
}

// GetSession retrieves the session repository for a given session ID from the Redis repository.
// It retrieves the session data from the repository, and unmarshals it into a models.Session.
// If the session data is not found in the repository, it returns an error indicating that the session was not found.
func (r *SessionsRepository) GetSession(ctx context.Context, sessionID string) (session models.Session, err error) {
	defer r.updateMetrics(&err, "GetSession")
	defer handleError(ctx, &err)
	defer r.logError(&err, "GetSession")

	body, err := r.rdb.Get(ctx, sessionID).Bytes()
	if err != nil {
		return
	}

	r.logger.Info("Unmarshal repository data")
	if err = json.Unmarshal(body, &session); err != nil {
		return
	}

	return
}

// UpdateLastActivityForSession updates the last activity time for a cached session in the Redis repository.
// It updates the LastActivity field of the cached session, and then caches the updated session.
func (r *SessionsRepository) UpdateLastActivityForSession(ctx context.Context,
	cachedSession *models.Session, lastActivityTime time.Time, ttl time.Duration) (err error) {
	defer r.updateMetrics(&err, "UpdateLastActivityForSession")
	defer handleError(ctx, &err)
	defer r.logError(&err, "UpdateLastActivityForSession")

	cachedSession.LastActivity = lastActivityTime
	err = r.SetSession(ctx, cachedSession, ttl)
	return
}

// GetSessionsForAccount retrieves the sessions associated with the specified account from the Redis repository.
func (r *SessionsRepository) GetSessionsForAccount(ctx context.Context,
	accountID string) (sessions map[string]*models.SessionInfo, err error) {
	defer r.updateMetrics(&err, "GetSessionsForAccount")
	defer handleError(ctx, &err)
	defer r.logError(&err, "GetSessionsForAccount")

	sessionsIds, err := r.rdb.SMembers(ctx, getKeyForAccountSessionsList(accountID)).Result()
	if err != nil {
		return
	}

	sessionsInfo, err := r.rdb.MGet(ctx, sessionsIds...).Result()
	if err != nil {
		return
	}

	sessions = make(map[string]*models.SessionInfo, len(sessionsInfo))
	for i := range sessionsInfo {
		var repository models.Session
		if sessionsInfo[i] == nil {
			continue
		}
		err = json.Unmarshal([]byte(sessionsInfo[i].(string)), &repository)
		if err != nil {
			return
		}

		sessions[repository.SessionID] = &models.SessionInfo{
			ClientIP:     repository.ClientIP,
			MachineID:    repository.MachineID,
			LastActivity: repository.LastActivity,
		}
	}

	return
}

type AccountSessions struct {
	Sessions []string `json:"sessions"`
}

func (r *SessionsRepository) TerminateAllSessions(ctx context.Context, accountID string) (err error) {
	defer r.updateMetrics(&err, "TerminateAllSessions")
	defer handleError(ctx, &err)
	defer r.logError(&err, "TerminateAllSessions")

	listKey := getKeyForAccountSessionsList(accountID)
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

func (r *SessionsRepository) logError(errptr *error, functionName string) {
	if errptr == nil || *errptr == nil {
		return
	}

	err := *errptr
	var repoErr = &models.ServiceError{}
	if errors.As(err, &repoErr) {
		r.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           repoErr.Msg,
				"error.code":          repoErr.Code,
			},
		).Error("sessions repository error occurred")
	} else {
		r.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("sessions repository error occurred")
	}
}

func (r *SessionsRepository) updateMetrics(errptr *error, functionName string) {
	if errptr == nil || *errptr == nil {
		r.metrics.IncCacheHits(functionName)
		return
	}
	if models.Code(*errptr) == models.NotFound {
		r.metrics.IncCacheMiss(functionName)
	}
}
