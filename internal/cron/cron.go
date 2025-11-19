package cron

import (
	"log"
	"time"

	"github.com/Wal-20/cli-chat-app/internal/services"
	"github.com/go-co-op/gocron"
)

func StartCronJobs() {
	s := gocron.NewScheduler(time.Local)

	s.Every(1).Day().Do(mainCleanup)
	s.StartAsync()
}

func mainCleanup() {
	cleanupNotifications()
	cleanupUserChatrooms()
}

func cleanupNotifications() {
	deleted, err := services.RemoveOldNotifications()
	if err != nil {
		log.Printf("Failed to remove old notifications: %v", err)
		return
	}
	log.Printf("Removed %v old notifications", deleted)
}

func cleanupUserChatrooms() {
	deleted, err := services.RemoveOldUserChatrooms()
	if err != nil {
		log.Printf("Failed to remove old user-chatroom associations: %v", err)
		return
	}
	log.Printf("Removed %v old user-chatroom associations", deleted)
}
