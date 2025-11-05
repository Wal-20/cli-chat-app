package handlers

import (
	"github.com/Wal-20/cli-chat-app/internal/repositories"
	"github.com/Wal-20/cli-chat-app/internal/services"
)

// Svcs holds initialized service singletons for handlers to use.
var Svcs struct {
	Auth         *services.AuthService
	Chat         *services.ChatroomService
	Message      *services.MessageService
	Notification *services.NotificationService
}

func InitHandlers() {
	userRepo := repositories.DefaultUserRepository()
	chatRepo := repositories.DefaultChatroomRepository()
	msgRepo := repositories.DefaultMessageRepository()

	Svcs.Auth = services.NewAuthService(userRepo)
	Svcs.Chat = services.NewChatroomService(chatRepo)
	Svcs.Message = services.NewMessageService(msgRepo, userRepo)
	Svcs.Notification = services.NewNotificationService()
}
