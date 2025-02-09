package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

	chatroomIDStr := r.PathValue("id")
	if chatroomIDStr == "" {
		http.Error(w, "No valid chatroom ID provided", http.StatusBadRequest)
		return
	}

	chatroomID, err := strconv.ParseUint(chatroomIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chatroom ID", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Content string `json:"content"`
	}

	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Unable to decode request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if requestBody.Content == "" || chatroomID == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	chatroomIdUint := uint(chatroomID)
	message := models.Message{
		UserId:  senderID,
		Content: requestBody.Content,
		ChatroomID: chatroomIdUint,
	}

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
	chatroomId := r.PathValue("id")
	isAdmin := r.Context().Value("isAdmin").(bool)

	if messageId == "" || chatroomId  == "" {
		http.Error(w, "Please provide a valid id for message and chatroom", http.StatusBadRequest)
		return
	}
	userId, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(w,"Unauthorized", http.StatusUnauthorized)
		return
	}

	var message models.Message
	if err := config.DB.First(&message, messageId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Cannot fetch message, not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching message", http.StatusInternalServerError)
		}
		return
	}

	if !isAdmin && message.UserId != userId {
		http.Error(w, "Cannot delete another user's message", http.StatusUnauthorized)
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

func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "not implemented yet",
	})
}
