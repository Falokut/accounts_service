package models

type SignInDTO struct {
	Email     string
	Password  string
	ClientIP  string
	MachineID string
}
