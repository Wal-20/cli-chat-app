package config

import (
	"gorm.io/driver/mysql"
	"os"
	"gorm.io/gorm"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/joho/godotenv"
	"log"
)

var DB *gorm.DB // global instance

func InitDB() error {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	dsn := os.Getenv("SERVICE_URI") 
	if dsn == "" {
		log.Fatal("CANNOT READ SERVICE_URI IN ENVIRONMENT")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Chatroom{},
		&models.UserChatroom{},
		&models.Message{},
		&models.Notification{},
	)

	if err != nil {
		return err
	}

	DB = db

	log.Println("DB SYNC")
	return nil
}

