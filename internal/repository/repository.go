package repository

import (
	"context"
	"time"

	"github.com/Falokut/accounts_service/internal/models"
)

type Transaction interface {
	Rollback() error
	Commit() error
}

// AccountRepository provides methods to interact with user accounts in the database.
//
//go:generate mockgen -source=repository.go -destination=mocks/accountRepository.go
type AccountRepository interface {
	// CreateAccount creates a new account in the database.
	CreateAccount(ctx context.Context, account models.Account) (Transaction, string, error)

	// IsAccountWithEmailExist checks if an account with the given email exists in the database.
	IsAccountWithEmailExist(ctx context.Context, email string) (bool, error)

	// GetAccountByEmail retrieves an account from the database using the email.
	GetAccountByEmail(ctx context.Context, email string) (models.Account, error)

	// GetCachedAccount retrieves the email using the account id.
	GetAccountEmail(ctx context.Context, accountId string) (string, error)

	// ChangePassword updates the password hash of an account with the given email in the database.
	ChangePassword(ctx context.Context, email string, passwordHash string) error

	// DeleteAccount removes the account with the given id from the database.
	DeleteAccount(ctx context.Context, id string) (Transaction, error)

	// Shutdown performs cleanup and shuts down the repository.
	Shutdown() error
}

// RegistrationRepository provides methods to interact with the registration repository.
//
//go:generate mockgen -source=repository.go -destination=mocks/RegistrationRepository.go
type RegistrationRepository interface {
	// IsAccountInRepository checks if the account associated with the given email is present in the repository.
	// It returns true if the account is found in the repository, otherwise false.
	// An error is returned if there is an issue while checking the repository.
	IsAccountExist(ctx context.Context, email string) (bool, error)

	// SetAccount caches the account with the given email and its details with a specified time-to-live duration.
	SetAccount(ctx context.Context, email string, account models.RegisteredAccount, ttl time.Duration) error

	// DeleteAccount removes the account with the given email from the repository.
	DeleteAccount(ctx context.Context, email string) error

	// GetAccount retrieves the cached account data using the email.
	GetAccount(ctx context.Context, email string) (models.RegisteredAccount, error)

	PingContext(ctx context.Context) error

	// Shutdown performs cleanup and shuts down the registration repository.
	Shutdown() error
}

// SessionsRepository provides methods to interact with the sessions repository.
//
//go:generate mockgen -source=repository.go -destination=mocks/SessionsRepository.go
type SessionsRepository interface {
	// CacheSession caches the session data.
	SetSession(ctx context.Context, session models.Session, ttl time.Duration) error

	// TerminateSessions terminates the specified sessions for the given account id.
	TerminateSessions(ctx context.Context, sessionsID []string, accountId string) error

	// UpdateLastActivityForSession updates the last activity time for the session.
	UpdateLastActivityForSession(ctx context.Context, session models.Session, lastActivityTime time.Time, ttl time.Duration) error

	// GetSession retrieves the cached session data for a specific session id.
	GetSession(ctx context.Context, sessionId string) (models.Session, error)

	// GetSessionsForAccount fetches the sessions associated with a particular account id.
	GetSessionsForAccount(ctx context.Context, accountId string) (map[string]models.SessionInfo, error)

	// GetSessionsIds fetches the sessions associated with a particular account id.
	GetSessionsIds(ctx context.Context, accountId string) ([]string, error)

	PingContext(ctx context.Context) error

	TerminateAllSessions(ctx context.Context, accountId string) error
	// Shutdown stops the repository operations.
	Shutdown() error
}

type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     string `yaml:"port" env:"DB_PORT"`
	Username string `yaml:"username" env:"DB_USERNAME"`
	Password string `yaml:"password" env:"DB_PASSWORD,env-required" env-default:"password"`
	DBName   string `yaml:"db_name" env:"DB_NAME"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE"`
}
