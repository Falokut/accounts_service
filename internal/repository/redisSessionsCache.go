package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Falokut/accounts_service/internal/model"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type redisSessionsCache struct {
	sessions_rdb         *redis.Client
	account_sessions_rdb *redis.Client
	logger               *logrus.Logger
	SessionTTL           time.Duration
}

// NewSessionCache creates a new session cache using the provided Redis options, logger, and session TTL.
// It initializes two Redis clients for session and account session caching, and verifies the connection to each Redis instance.
func NewSessionCache(sessionCacheOpt *redis.Options, accountSessionsOpt *redis.Options, logger *logrus.Logger, sessionTTL time.Duration) (*redisSessionsCache, error) {
	logger.Infoln("Creating session cache client")

	// Initialize a Redis client for session caching
	sessions_rdb := redis.NewClient(sessionCacheOpt)
	if sessions_rdb == nil {
		return nil, errors.New("can't create new redis client")
	}

	logger.Infoln("Pinging session cache client")
	// Verify the connection to the session cache Redis instance
	_, err := sessions_rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("connection is not established: %s", err.Error())
	}

	logger.Infoln("Creating account sessions cache client")
	// Initialize a Redis client for account sessions caching
	account_sessions_rdb := redis.NewClient(accountSessionsOpt)
	if account_sessions_rdb == nil {
		return nil, errors.New("can't create new redis client")
	}

	logger.Infoln("Pinging session cache client")
	// Verify the connection to the account sessions cache Redis instance
	_, err = account_sessions_rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("connection is not established: %s", err.Error())
	}

	// Return the initialized redisSessionsCache with the configured Redis clients, logger, and session TTL
	return &redisSessionsCache{sessions_rdb: sessions_rdb, account_sessions_rdb: account_sessions_rdb, logger: logger, SessionTTL: sessionTTL}, nil
}

func (r *redisSessionsCache) PingContext(ctx context.Context) error {
	if err := r.sessions_rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error while pinging sessions cache: %w", err)
	}
	if err := r.account_sessions_rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error while pinging account sessions cache: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the token cache repository by closing the Redis client for session caching.
func (r *redisSessionsCache) Shutdown() error {
	r.logger.Infoln("Token cache repository shutting down")
	return r.sessions_rdb.Close()
}

func (r *redisSessionsCache) CacheSession(ctx context.Context, toCache model.SessionCache) error {
	// Create a new span for caching the session data
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.CacheSession")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil)

	// Log a message indicating the marshalling of the data
	r.logger.Info("Marshalling data")
	// Marshal the session data into a JSON format
	serialized, err := json.Marshal(toCache)
	if err != nil {
		return err
	}

	// Log a message indicating the caching of the session data
	r.logger.Info("Caching sessions data")
	// Cache the serialized session data in Redis with the specified TTL
	_, err = r.sessions_rdb.Set(ctx, toCache.SessionID, serialized, r.SessionTTL).Result()
	if err != nil {
		return err
	}

	// Cache the account session data
	if err = r.cacheAccountSession(ctx, toCache); err != nil {
		// If an error occurs, delete the session data from the cache
		r.sessions_rdb.Del(ctx, toCache.SessionID)
		return err
	}

	return nil
}

func (r *redisSessionsCache) TerminateSessions(ctx context.Context, sessionsID []string, accountID string) error {
	// Create a new span for terminating sessions
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.TerminateSessions")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil)

	// Get the list of sessions associated with the specified account
	sessions, err := r.GetSessionsList(ctx, accountID)
	if err != nil {
		return err
	}

	// Iterate through the session IDs and delete them from the cache
	if err := r.sessions_rdb.Del(ctx, sessionsID...).Err(); err != nil {
		return err
	}
	for _, sessionID := range sessionsID {
		r.logger.Debugf("Deleting session with id %s", sessionID)
		for i, id := range sessions {
			if id == sessionID {
				sessions[i] = sessions[len(sessions)-1]
				sessions = sessions[:len(sessions)-1]
				break
			}
		}
	}

	// Update the account's remaining sessions
	err = r.UpdateSessionsForAccount(ctx, AccountSessions{Sessions: sessions}, accountID)

	return err
}

func (r *redisSessionsCache) GetSessionsList(ctx context.Context, accountID string) ([]string, error) {
	// Start a new span for tracing the GetSessionsList operation
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.GetSessionsList")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil && !errors.Is(err, redis.Nil))

	res, err := r.account_sessions_rdb.Get(ctx, accountID).Bytes()
	if errors.Is(err, redis.Nil) {
		return []string{}, nil // No error, just an empty sessions map for the account
	}

	var sessions AccountSessions
	if err = json.Unmarshal(res, &sessions); err != nil {
		return []string{}, err
	}

	return sessions.Sessions, nil
}

// cacheAccountSession caches the session information for a specific account in the Redis cache.
// It retrieves the existing session data, updates it with new information, and then updates the cache.
// This function is a part of the RedisSessionsCache struct.
func (r *redisSessionsCache) cacheAccountSession(ctx context.Context, toCache model.SessionCache) error {
	// Start a new span for tracing the cacheAccountSession operation
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.cacheAccountSession")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil)
	// Attempt to retrieve the existing session data for the account from the Redis cache
	body, err := r.account_sessions_rdb.Get(ctx, toCache.AccountID).Bytes()
	if err != nil && err != redis.Nil {
		return err
	}

	// Initialize a struct to hold the session cache data
	var sessionsCache AccountSessions
	if err != redis.Nil {
		// Log an informational message indicating unmarshaling of data
		r.logger.Info("Unmarshal data")
		// Unmarshal the retrieved data if it exists
		if err = json.Unmarshal(body, &sessionsCache); err != nil {
			return err
		}
	}
	sessionsCache.Sessions = append(sessionsCache.Sessions, toCache.SessionID)

	// Update the sessions for the account in the Redis cache
	return r.UpdateSessionsForAccount(ctx, sessionsCache, toCache.AccountID)
}

// GetSessionCache retrieves the session cache for a given session ID from the Redis cache.
// It starts a new span for tracing, retrieves the session data from the cache, and unmarshals it into a model.SessionCache.
// If the session data is not found in the cache, it returns an error indicating that the session was not found.
func (r *redisSessionsCache) GetSessionCache(ctx context.Context, sessionID string) (model.SessionCache, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.GetSessionCache")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil)

	// Retrieve the session data for the given session ID from the Redis cache
	body, err := r.sessions_rdb.Get(ctx, sessionID).Bytes()
	if errors.Is(err, redis.Nil) {
		return model.SessionCache{}, ErrSessionNotFound
	} else if err != nil {
		return model.SessionCache{}, err
	}

	// Unmarshal the retrieved cache data into a model.SessionCache struct
	r.logger.Info("Unmarshal cache data")
	var session model.SessionCache
	if err = json.Unmarshal(body, &session); err != nil {
		return model.SessionCache{}, err
	}

	return session, nil
}

// UpdateLastActivityForSession updates the last activity time for a cached session in the Redis cache.
// It starts a new span for tracing, updates the LastActivity field of the cached session, and then caches the updated session.
func (r *redisSessionsCache) UpdateLastActivityForSession(ctx context.Context,
	cachedSession model.SessionCache, sessionID string, lastActivityTime time.Time) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.UpdateLastActivityForSession")
	defer span.Finish()
	// Update the LastActivity field of the cached session with the provided LastActivityTime
	cachedSession.LastActivity = lastActivityTime
	err := r.CacheSession(ctx, cachedSession)
	span.SetTag("error", err != nil)
	return err
}

type SessionInfo struct {
	ClientIP     string    `json:"client_ip"`
	MachineID    string    `json:"machine_id"`
	LastActivity time.Time `json:"last_activity"`
}

type AccountSessionsCacheModel struct {
	Sessions map[string]SessionInfo `json:"sessions"`
}

// GetSessionsForAccount retrieves the sessions associated with the specified account from the Redis cache.
func (r *redisSessionsCache) GetSessionsForAccount(ctx context.Context, accountID string) (map[string]SessionInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.GetSessionsForAccount")
	defer span.Finish()
	var err error
	defer span.SetTag("error", err != nil && !errors.Is(err, redis.Nil))

	// Retrieve the cached data for the specified accountID from the Redis database.
	body, err := r.account_sessions_rdb.Get(ctx, accountID).Bytes()
	if errors.Is(err, redis.Nil) {
		return map[string]SessionInfo{}, ErrSessionNotFound
	}

	var sessions AccountSessions
	if err = json.Unmarshal(body, &sessions); err != nil {
		return map[string]SessionInfo{}, err
	}

	var SessionsInfo AccountSessionsCacheModel
	SessionsInfo.Sessions = make(map[string]SessionInfo, len(sessions.Sessions))
	sessionsInfo, err := r.sessions_rdb.MGet(ctx, sessions.Sessions...).Result()
	if err != nil {
		return map[string]SessionInfo{}, err
	}

	for i, sessionInfo := range sessionsInfo {
		var cache model.SessionCache
		err = json.Unmarshal([]byte(sessionInfo.(string)), &cache)
		if err != nil {
			r.logger.Error(err)
			return map[string]SessionInfo{}, err
		}

		SessionsInfo.Sessions[sessions.Sessions[i]] = SessionInfo{
			ClientIP:     cache.ClientIP,
			MachineID:    cache.MachineID,
			LastActivity: cache.LastActivity,
		}
	}

	return SessionsInfo.Sessions, nil
}

type AccountSessions struct {
	Sessions []string `json:"sessions"`
}

// UpdateSessionsForAccount updates the sessions associated with the specified account in the Redis cache.
func (r *redisSessionsCache) UpdateSessionsForAccount(ctx context.Context, sessions AccountSessions, accountID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.UpdateSessionsForAccount")
	defer span.Finish()
	var err error
	defer span.SetTag("error", err != nil && !errors.Is(err, redis.Nil))

	// If the sessions map is empty, remove the corresponding entry from the Redis database.
	if len(sessions.Sessions) == 0 {
		err := r.account_sessions_rdb.Del(ctx, accountID).Err()
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	// Marshal the sessions data into a JSON object.
	r.logger.Info("Marshalling data")
	serialized, err := json.Marshal(sessions)
	if err != nil {
		return err
	}

	// Store the serialized sessions data in the Redis database with the specified accountID.
	r.logger.Info("Caching data")
	_, err = r.account_sessions_rdb.Set(ctx, accountID, serialized, r.SessionTTL).Result()

	return err
}

func (r *redisSessionsCache) TerminateAllSessions(ctx context.Context, accountID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionsCache.TerminateAllSessions")
	defer span.Finish()
	var err error
	defer span.SetTag("error", err != nil && !errors.Is(err, redis.Nil))

	body, err := r.account_sessions_rdb.Get(ctx, accountID).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil // No error, just an empty sessions map for the account
	}

	var sessions AccountSessions
	if err = json.Unmarshal(body, &sessions); err != nil {
		return err
	}

	if err = r.account_sessions_rdb.Del(ctx, sessions.Sessions...).Err(); err != nil {
		return err
	}
	if err = r.sessions_rdb.Del(ctx, sessions.Sessions...).Err(); err != nil {
		return err
	}

	return nil
}
