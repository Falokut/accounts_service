package models

import (
	"time"
)

type Account struct {
	Id               string    `db:"id" json:"id"`
	Email            string    `db:"email" json:"email"`
	Password         string    `db:"password_hash" json:"-"`
	RegistrationDate time.Time `db:"registration_date" json:"registration_date"`
}

// RegisteredAccount represents the account data in the registration database.
type RegisteredAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
