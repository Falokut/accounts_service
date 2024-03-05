package models

import "time"

type SessionInfo struct {
	ClientIP     string    `json:"client_ip"`
	MachineID    string    `json:"machine_id"`
	LastActivity time.Time `json:"last_activity"`
}
