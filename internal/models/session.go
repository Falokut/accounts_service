package models

import "time"

type Session struct {
	SessionId    string    `json:"session_id"`
	AccountId    string    `json:"account_id"`
	MachineId    string    `json:"machine_id"`
	ClientIp     string    `json:"client_ip"`
	LastActivity time.Time `json:"last_activity"`
}
