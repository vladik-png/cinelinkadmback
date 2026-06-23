package models

import "time"

type EmployeeStatus struct {
	EmployeeID uint      `gorm:"primaryKey"`
	IsOnline   bool      `gorm:"default:false"`
	LastSeen   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
