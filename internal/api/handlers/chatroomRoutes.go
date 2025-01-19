package handlers
import (
	"encoding/json"
	"net/http"
    "gorm.io/gorm"
	"errors"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/config"
)


func GetChatrooms(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	id := r.URL.Query().Get("id")

	if id == "" {
		var chatrooms []models.Chatroom
		result := config.DB.Find(&chatrooms)

		if result.Error != nil {
			http.Error(w, "Failed to retrieve chatrooms", http.StatusInternalServerError)
			return
		}

		encoder.Encode(map[string]interface{}{
			"Chatrooms": chatrooms,
		})

	} else {

		var chatroom models.Chatroom
		result := config.DB.First(&chatroom, id)

		if result.Error != nil {
			http.Error(w, "Failed to retrieve chatroom", http.StatusInternalServerError)
			return
		}
		
		encoder.Encode(map[string]interface{}{
			"Chatroom": chatroom,
		})
	}
}

func GetUsersByChatroom(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	chatroomId := r.PathValue("id")

	if chatroomId == "" {
		http.Error(w, "Please provide a valid ID", http.StatusBadRequest)
		return
	}

	var chatroom models.Chatroom
	result := config.DB.Preload("Users").First(&chatroom, chatroomId)
	if result.Error != nil {
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}

	encoder.Encode(map[string]interface{}{
		"Users": chatroom.Users,
	})

}

func GetMessagesByChatroom(w http.ResponseWriter, r *http.Request) {
	chatroomId := r.PathValue("id")
	encoder := json.NewEncoder(w)

	if chatroomId == "" {
		http.Error(w, "Please provide a valid ID", http.StatusBadRequest)
		return
	}

	searchTerms := r.URL.Query().Get("search")
	var messages []models.Message

	// Build the query
	query := config.DB.Where("chatroom_id = ?", chatroomId)

	// Add search terms if provided
	if searchTerms != "" {
		query = query.Where("content LIKE ?", "%"+searchTerms+"%")
	}

	// Execute the query
	err := query.Order("created_at ASC").Limit(20).Find(&messages).Error

	if err != nil {
		http.Error(w, "No messages found", http.StatusNotFound)
		return
	}

	encoder.Encode(map[string]interface{}{
		"Messages": messages,
	})
}


func CreateChatroom(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("userID").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized: missing or invalid user ID", http.StatusUnauthorized)
		return
	}

	// Parse the request body to extract the recipient ID
	var requestBody struct {
		RecipientID uint   `json:"recipient_id"`
		Title       string `json:"title"`
		// MaxUserCount uint `json:"maxUserCount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate the recipient ID
	if requestBody.RecipientID == 0 {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}
	if requestBody.RecipientID == userID {
		http.Error(w, "Cannot create a chatroom with yourself", http.StatusBadRequest)
		return
	}
	
	userIDs := []uint{userID, requestBody.RecipientID}

	// Query to find users by IDs
	var users []models.User
	if err := config.DB.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		http.Error(w, "Error retrieving Users", http.StatusConflict)
	}

	// Create a new chat room
	newChatRoom := models.Chatroom {
		Users:       users,
		AdminId: userID,
		IsPublic: true,
		Title: requestBody.Title,
	}

	// Save the new chat room to the database
	if err := config.DB.Create(&newChatRoom).Error; err != nil {
		http.Error(w, "Failed to create chat room", http.StatusInternalServerError)
		return
	}

	// Respond with the created chat room details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newChatRoom)
}

func deleteChatroom(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	encoder := json.NewEncoder(w)

	if id == "" {
		http.Error(w, "Please provide a valid ID", http.StatusBadRequest)
		return
	}

	var chatroom models.Chatroom
	if err := config.DB.First(&chatroom, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Chatroom not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving chatroom", http.StatusInternalServerError)
		}
		return
	}

	if err := config.DB.Delete(&chatroom).Error; err != nil {
		http.Error(w, "Error deleting chatroom", http.StatusInternalServerError)
		return
	}

	encoder.Encode(map[string]interface{}{
		"Status": "Deleted chatroom successfully",
	})
}

func JoinChatroom(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	userID, ok := r.Context().Value("userID").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized: missing or invalid user ID", http.StatusUnauthorized)
		return
	}

	chatroomID := r.PathValue("id")
	if chatroomID == "" {
		http.Error(w, "Please provide a valid chatroom ID", http.StatusBadRequest)
		return
	}

	var chatroom models.Chatroom
	result := config.DB.Preload("Users").First(&chatroom, chatroomID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Chatroom not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving chatroom", http.StatusInternalServerError)
		}
		return
	}

	var user models.User
	if result = config.DB.First(&user, userID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		}
		return
	}

	// Check if user is already a member
	for _, u := range chatroom.Users {
		if u.ID == user.ID {
			http.Error(w, "User is already a member of the chatroom", http.StatusConflict)
			return
		}
	}

	// use OTP if chatroom isn't public

	chatroom.Users = append(chatroom.Users, user)

	if err := config.DB.Save(&chatroom).Error; err != nil {
		http.Error(w, "Error saving chatroom", http.StatusInternalServerError)
		return
	}

	encoder.Encode(map[string]interface{}{
		"Status":   "User added to chatroom successfully",
		"Chatroom": chatroom,
	})
}

