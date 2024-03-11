package handler

import (
	"errors"
	"net/mail"

	accounts_service "github.com/Falokut/accounts_service/pkg/accounts_service/v1/protos"
)

func validateSignupInput(input *accounts_service.CreateAccountRequest) error {
	if input == nil {
		return errors.New("request body not valid")
	}

	if input.Password != input.RepeatPassword {
		return errors.New("passwords don't match")
	}

	if err := validateUsername(input.Username); err != nil {
		return err
	}

	if err := validatePassword(input.Password); err != nil {
		return err
	}
	if err := validateEmail(input.Email); err != nil {
		return err
	}

	return nil
}
func validateUsername(username string) error {
	usernameLength := len(username)
	if usernameLength < 3 || usernameLength > 32 {
		return errors.New("username must be less than 32 symbols and more than 3")
	}

	return nil
}

func validatePassword(password string) error {
	passwordLength := len(password)
	if passwordLength < 6 || passwordLength > 32 {
		return errors.New("password must be less than 32 symbols and more than 6")
	}

	return nil
}

func validateEmail(email string) error {
	if len(email) > 100 || len(email) < 4 {
		return errors.New("email must be less than 100 symbols and more than 4")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}
