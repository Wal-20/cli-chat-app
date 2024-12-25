package main

import (
	"log"
	"os"
	"github.com/Wal-20/cli-chat-app/internal/api"
	"github.com/Wal-20/cli-chat-app/internal/api/handlers"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"gorm.io/driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)


func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("SERVICE_URI") 
	if dsn == "" {
		log.Fatal("CANNOT READ SERVICE_URI IN ENVIRONMENT")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("failed to migrate Database", err)
	}

	log.Println("DB SYNC")

	handlers.InitializeDB(db)

	// Initialize server
	api.NewServer()
}

