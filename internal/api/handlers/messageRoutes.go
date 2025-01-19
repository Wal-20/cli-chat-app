package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"gorm.io/gorm"
)

func SendMessage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)
	senderID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(w,"Unauthorized", http.StatusUnauthorized)
		return
	}
	var message models.Message
	err := decoder.Decode(&message)

	if err != nil {
		http.Error(w, "Unable to decode request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	if message.Content == "" || message.ChatroomID == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	message.UserId = senderID

	result := config.DB.Create(&message)

	if result.Error != nil {
		http.Error(w, "Unable to create messsage", http.StatusInternalServerError)
		return
	}

	encoder.Encode(map[string]interface{}{
		"Status": "success",
		"Message": message,
	})
}


func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageId := r.PathValue("messageId")

	if messageId == "" {
		http.Error(w, "Please provide a valid id", http.StatusBadRequest)
		return
	}
	
	if result := config.DB.Delete(messageId); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Message not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Error deleting message", http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "Message Deleted Successfully",
	})
}

