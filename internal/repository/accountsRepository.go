package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Falokut/accounts_service/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
)

const (
	accountTableName = "accounts"
)

type accountsRepository struct {
	db *sqlx.DB
}

// NewPostgreDB creates a new connection to the PostgreSQL database.
func NewPostgreDB(cfg DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))

	return db, err
}

// NewAccountsRepository creates a new instance of the accountsRepository using the provided database connection.
func NewAccountsRepository(db *sqlx.DB) *accountsRepository {
	return &accountsRepository{db: db}
}

// Shutdown closes the database connection.
func (r *accountsRepository) Shutdown() error {
	return r.db.Close()
}

// CreateAccount creates a new account in the database.
func (r *accountsRepository) CreateAccount(ctx context.Context, account model.Account) (*sql.Tx, string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx,
		"accountsRepository.CreateAccount")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil)

	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, registration_date) VALUES ($1, $2, $3) RETURNING id;", accountTableName)
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, "", err
	}

	row := tx.QueryRowContext(ctx, query, account.Email, account.Password, account.RegistrationDate)

	var id string
	if err = row.Scan(&id); err != nil {
		tx.Rollback()
		return nil, "", err
	}

	return tx, id, nil
}

// IsAccountWithEmailExist checks if an account with the given email exists in the database.
// It returns a boolean indicating the existence and an error, if any.
func (r *accountsRepository) IsAccountWithEmailExist(ctx context.Context, email string) (bool, error) {
	// Start a new span for tracing.
	span, _ := opentracing.StartSpanFromContext(ctx, "accountsRepository.IsAccountWithEmailExist")

	defer span.Finish() // Finish the span when the function ends.

	var err error
	defer span.SetTag("error", err != nil && !errors.Is(err, sql.ErrNoRows))

	// Prepare the SQL query to check for the existence of the account.
	query := fmt.Sprintf("SELECT id FROM %s WHERE email=$1 LIMIT 1;", accountTableName)

	var UUID string
	// Execute the query to check for the existence of the account with the given email.
	err = r.db.GetContext(ctx, &UUID, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // Return false if no rows were found (account does not exist).
	}
	if err != nil {
		return false, err // Return false and the error if an error other than sql.ErrNoRows occurs.
	}

	return true, nil // If no error occurred, the account exists.
}

// GetAccountByEmail retrieves a account from the database based on the provided email.
// It returns the retrieved account and an error, if any.
func (r *accountsRepository) GetAccountByEmail(ctx context.Context, email string) (model.Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "accountsRepository.GetAccountByEmail")
	defer span.Finish()

	var err error
	defer span.SetTag("error", err != nil)

	// Prepare the SQL query to retrieve the user account based on the provided email.
	query := fmt.Sprintf("SELECT * FROM %s WHERE email=$1 LIMIT 1;", accountTableName)

	var acc model.Account
	// Execute the query to retrieve the user account.
	err = r.db.GetContext(ctx, &acc, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Account{}, errors.New("user with this email doesn't exist") // Return an error if no account was found with the provided email.
	}

	return acc, err // Return the retrieved account and any error that occurred during retrieval.
}

// ChangePassword updates the password hash of an account with the given email in the database.
// It takes the email and the new password hash as input and returns an error, if any.
func (r *accountsRepository) ChangePassword(ctx context.Context, email string, passwordHash string) error {
	// Start a new span for tracing.
	span, _ := opentracing.StartSpanFromContext(ctx, "accountsRepository.ChangePassword")

	defer span.Finish() // Finish the span when the function ends.

	var err error
	defer span.SetTag("error", err != nil)

	// Prepare the SQL query to update the password hash of the account with the given email.
	query := fmt.Sprintf("UPDATE %s SET password_hash=$1 WHERE email=$2;", accountTableName)

	// Execute the query to update the password hash.
	res, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return err // Return the error if the query execution fails.
	}

	// Get the number of rows affected by the update.
	num, err := res.RowsAffected()
	if err != nil || num == 0 {
		return errors.New("rows are not affected") // Return an error if no rows are affected by the update.
	}

	return nil // Return nil to indicate success.
}

// DeleteAccount deletes the account with the given ID from the database.
// It takes the ID of the account as input and returns an error, if any.
func (r *accountsRepository) DeleteAccount(ctx context.Context, accountID string) (*sql.Tx, error) {
	// Start a new span for tracing.
	span, _ := opentracing.StartSpanFromContext(ctx, "accountsRepository.DeleteAccount")
	defer span.Finish() // Finish the span when the function ends.
	var err error
	defer span.SetTag("error", err != nil)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	// Prepare the SQL query to delete the account with the given ID.
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1;", accountTableName)

	// Execute the query to delete the account.
	_, err = tx.ExecContext(ctx, query, accountID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}
