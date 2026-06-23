package main

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type EmployeeStatus struct {
	EmployeeID uint      `gorm:"primaryKey"`
	IsOnline   bool      `gorm:"default:false"`
	LastSeen   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func main() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&EmployeeStatus{})

	status := EmployeeStatus{
		EmployeeID: 1,
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Save(&status)

	var count int64
	db.Model(&EmployeeStatus{}).Count(&count)
	fmt.Printf("Count after Save: %d\n", count)
}
