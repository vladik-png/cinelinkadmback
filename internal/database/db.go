package database

import (
	"log"

	"admin-aws/internal/models"
	"admin-aws/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var ChatDB *gorm.DB

func InitDB() {
	dsn := config.GetEnv("DATABASE_URL", "")
	chatDsn := config.GetEnv("CHAT_DATABASE_URL", "")

	if dsn == "" {
		log.Println("DATABASE_URL not found, DB disabled")
	} else {
		var err error
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		DB.AutoMigrate(
			&models.ServerLog{}, 
			&models.ServerAlert{},
		)
		log.Println("Database initialized successfully")
	}

	if chatDsn == "" {
		log.Println("CHAT_DATABASE_URL not found, Chat DB disabled")
	} else {
		var err error
		ChatDB, err = gorm.Open(postgres.Open(chatDsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to chat database:", err)
		}

		ChatDB.AutoMigrate(
			&models.CorporateChat{},
			&models.CorporateChatMember{},
			&models.CorporateChatMessage{},
			&models.EmployeeStatus{},
		)
		log.Println("Chat database initialized successfully")
	}
}

func LogEvent(serverID, action, status, details string) {
	if DB == nil { return }
	entry := models.ServerLog{
		ServerID: serverID,
		Action:   action,
		Status:   status,
		Details:  details,
	}
	DB.Create(&entry)
	log.Printf("[%s] %s - %s: %s\n", status, serverID, action, details)
}

func CreateAlert(serverID, alertType, message string) {
	if DB == nil { return }
	var count int64
	DB.Model(&models.ServerAlert{}).Where("server_id = ? AND type = ? AND resolved = ?", serverID, alertType, false).Count(&count)
	
	if count == 0 {
		alert := models.ServerAlert{
			ServerID: serverID,
			Type:     alertType,
			Message:  message,
		}
		DB.Create(&alert)
		log.Printf("ALERT [%s]: %s - %s\n", alertType, serverID, message)
	}
}
