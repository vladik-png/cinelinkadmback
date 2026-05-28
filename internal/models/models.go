package models

import "time"

type ServerLog struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
	ServerID  string    `gorm:"index"`
	Component string    
	Action    string    
	Status    string    
	Details   string    `gorm:"type:text"`
}

type ServerAlert struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
	ServerID  string    `gorm:"index"`
	Type      string
	Message   string
	Resolved  bool      `gorm:"default:false"`
}

type ServerConfig struct {
	ID         string
	Provider   string
	Platform   string
	MacAddress string
	AgentURL   string
	WoLTargets []string
}

type ServerState struct {
	LastSeen time.Time
	Metrics  map[string]interface{}
}
