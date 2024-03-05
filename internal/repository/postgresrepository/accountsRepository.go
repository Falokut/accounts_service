package postgresrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Falokut/accounts_service/internal/models"
	"github.com/Falokut/accounts_service/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	accountTableName = "accounts"
)

type AccountsRepository struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

// NewPostgreDB creates a new connection to the PostgreSQL database.
func NewPostgreDB(cfg *repository.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))

	return db, err
}

// NewAccountsRepository creates a new instance of the accountsRepository using the provided database connection.
func NewAccountsRepository(db *sqlx.DB, logger *logrus.Logger) *AccountsRepository {
	return &AccountsRepository{db: db, logger: logger}
}

// Shutdown closes the database connection.
func (r *AccountsRepository) Shutdown() {
	r.logger.Info("Acounts repository shutting down")

	err := r.db.Close()
	if err != nil {
		r.logger.Errorf("error while shutting down accounts repository %v", err)
	}
}

func (r *AccountsRepository) PingContext(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// CreateAccount creates a new account in the database.
func (r *AccountsRepository) CreateAccount(ctx context.Context,
	account models.Account) (restx repository.Transaction, id string, err error) {
	defer r.handleError(ctx, &err, "CreateAccount")

	query := fmt.Sprintf(`INSERT INTO %s (email, password_hash, registration_date)
        VALUES ($1, $2, $3) RETURNING id;`, accountTableName)
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	row := tx.QueryRowContext(ctx, query, account.Email, account.Password, account.RegistrationDate)
	if err = row.Scan(&id); err != nil {
		err = tx.Rollback()
		return
	}

	return tx, id, nil
}

// IsAccountWithEmailExist checks if an account with the given email exists in the database.
// It returns a boolean indicating the existence and an error, if any.
func (r *AccountsRepository) IsAccountWithEmailExist(ctx context.Context, email string) (exist bool, err error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE email=$1 LIMIT 1;", accountTableName)

	var id string
	err = r.db.GetContext(ctx, &id, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		r.handleError(ctx, &err, "IsAccountWithEmailExist")
		return
	}

	return true, nil
}

// GetAccountByEmail retrieves a account from the database based on the provided email.
// It returns the retrieved account and an error, if any.
func (r *AccountsRepository) GetAccountByEmail(ctx context.Context, email string) (account models.Account, err error) {
	defer r.handleError(ctx, &err, "GetAccountByEmail")

	query := fmt.Sprintf("SELECT * FROM %s WHERE email=$1 LIMIT 1;", accountTableName)

	err = r.db.GetContext(ctx, &account, query, email)
	return
}

// ChangePassword updates the password hash of an account with the given email in the database.
// It takes the email and the new password hash as input and returns an error, if any.
func (r *AccountsRepository) ChangePassword(ctx context.Context, email, passwordHash string) (err error) {
	defer r.handleError(ctx, &err, "ChangePassword")

	query := fmt.Sprintf("UPDATE %s SET password_hash=$1 WHERE email=$2;", accountTableName)

	res, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return
	}

	num, err := res.RowsAffected()
	if err != nil || num == 0 {
		err = errors.New("rows are not affected")
		return
	}

	return
}

// DeleteAccount deletes the account with the given ID from the database.
// It takes the ID of the account as input and returns an error, if any.
func (r *AccountsRepository) DeleteAccount(ctx context.Context, accountID string) (restx repository.Transaction, err error) {
	defer r.handleError(ctx, &err, "DeleteAccount")

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1", accountTableName)

	_, err = tx.ExecContext(ctx, query, accountID)
	if err != nil {
		err = tx.Rollback()
		return
	}
	return tx, nil
}

func (r *AccountsRepository) GetAccountEmail(ctx context.Context, accountID string) (email string, err error) {
	defer r.handleError(ctx, &err, "GetAccountEmail")

	query := fmt.Sprintf("SELECT email FROM %s WHERE id=$1", accountTableName)
	err = r.db.GetContext(ctx, &email, query, accountID)
	if err != nil {
		return
	}
	return
}

func (r *AccountsRepository) handleError(ctx context.Context, err *error, functionName string) {
	if ctx.Err() != nil {
		var code models.ErrorCode
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			code = models.Canceled
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			code = models.DeadlineExceeded
		}
		*err = models.Error(code, ctx.Err().Error())
		r.logError(*err, functionName)
		return
	}

	if err == nil || *err == nil {
		return
	}

	r.logError(*err, functionName)
	var repoErr = &models.ServiceError{}
	if !errors.As(*err, &repoErr) {
		switch {
		case errors.Is(*err, sql.ErrNoRows):
			*err = models.Error(models.NotFound, "account not found")
		case *err != nil:
			*err = models.Error(models.Internal, "repository internal error")
		}
	}
}

func (r *AccountsRepository) logError(err error, functionName string) {
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
		).Error("accounts repository error occurred")
	} else {
		r.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("accounts repository error occurred")
	}
}
