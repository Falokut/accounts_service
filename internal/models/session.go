package models

import "time"

type Session struct {
	SessionID    string    `json:"session_id"`
	AccountID    string    `json:"account_id"`
	MachineID    string    `json:"machine_id"`
	ClientIP     string    `json:"client_ip"`
	LastActivity time.Time `json:"last_activity"`
}
