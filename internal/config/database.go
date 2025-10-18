package config

import (
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB // global instance

func InitDB() error {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, checking environment variables instead...")
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
