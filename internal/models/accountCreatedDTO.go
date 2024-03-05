package models

import "time"

type AccountCreatedDTO struct {
	ID               string    `db:"id" json:"id"`
	Username         string    `json:"username"`
	Email            string    `db:"email" json:"email"`
	RegistrationDate time.Time `db:"registration_date" json:"registration_date"`
}
