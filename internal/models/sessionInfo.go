package models

import "time"

type SessionInfo struct {
	ClientIp     string    `json:"client_ip"`
	MachineId    string    `json:"machine_id"`
	LastActivity time.Time `json:"last_activity"`
}
