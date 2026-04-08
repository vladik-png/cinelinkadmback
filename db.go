package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServerLog struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
	ServerID  string    `gorm:"index"`
	Action    string
	Status    string
	Details   string
}

type ServerAlert struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
	ServerID  string    `gorm:"index"`
	Type      string
	Message   string
	Resolved  bool      `gorm:"default:false"`
}

var db *gorm.DB

func initDB() {

	dsn := getEnv("DATABASE_URL", "")
	
	if dsn == "" {
		log.Fatal("DATABASE_URL не знайдено у файлі .env!")
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(" Failed to connect to NeonDB:", err)
	}

	db.AutoMigrate(&ServerLog{}, &ServerAlert{})
	log.Println("NeonDB (PostgreSQL) initialized successfully")
}

func logEvent(serverID, action, status, details string) {
	entry := ServerLog{
		ServerID: serverID,
		Action:   action,
		Status:   status,
		Details:  details,
	}
	db.Create(&entry)
	log.Printf("[%s] %s - %s: %s\n", status, serverID, action, details)
}

func createAlert(serverID, alertType, message string) {
	var count int64
	db.Model(&ServerAlert{}).Where("server_id = ? AND type = ? AND resolved = ?", serverID, alertType, false).Count(&count)
	
	if count == 0 {
		alert := ServerAlert{
			ServerID: serverID,
			Type:     alertType,
			Message:  message,
		}
		db.Create(&alert)
		log.Printf("ALERT [%s]: %s - %s\n", alertType, serverID, message)
	}
}