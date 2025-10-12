package handlers

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/Wal-20/cli-chat-app/internal/config"
    "github.com/Wal-20/cli-chat-app/internal/models"
    "github.com/Wal-20/cli-chat-app/internal/utils"
    "gorm.io/gorm"
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

func GetPublicChatrooms(w http.ResponseWriter, r *http.Request) {

    userID, ok := r.Context().Value("userID").(uint)
    if !ok || userID == 0 {
        http.Error(w, "Unauthorized: missing or invalid user ID", http.StatusUnauthorized)
        return
    }

    encoder := json.NewEncoder(w)
    chatrooms, err := Svcs.Chat.GetPublicChatrooms(userID)
    if err != nil {
        http.Error(w, "Failed to retrieve chatrooms", http.StatusInternalServerError)
        return
    }
    encoder.Encode(map[string]interface{}{"Chatrooms": chatrooms})

}

func GetUsersByChatroom(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	chatroomId := r.PathValue("id")

	if chatroomId == "" {
		http.Error(w, "Please provide a valid ID", http.StatusBadRequest)
		return
	}

	// Fetch all user-chatroom associations for the chatroom
	var userChatrooms []models.UserChatroom
	if err := config.DB.Where("chatroom_id = ?", chatroomId).Find(&userChatrooms).Error; err != nil {
		http.Error(w, "Error fetching user-chatroom associations", http.StatusInternalServerError)
		return
	}

    // Match client expectation key: userChatroom
    encoder.Encode(map[string]interface{}{
        "userChatroom": userChatrooms,
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
	var messages []models.MessageWithUser

	query := config.DB.
		Table("messages").
		Select("messages.content, messages.created_at, users.name AS username").
		Joins("JOIN users ON messages.user_id = users.id").
		Where("messages.chatroom_id = ?", chatroomId)

	if searchTerms != "" {
		query = query.Where("messages.content LIKE ?", "%"+searchTerms+"%")
	}

	err := query.Order("messages.created_at ASC").Limit(20).Scan(&messages).Error
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

	// Parse request body
	var requestBody struct {
		RecipientID  uint   `json:"recipient_id"`
		Title        string `json:"title"`
		MaxUserCount uint   `json:"maxUserCount"`
		IsPublic     bool   `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input (recipient optional)
	if requestBody.RecipientID == userID && requestBody.RecipientID != 0 {
		http.Error(w, "Cannot create a chatroom with yourself", http.StatusBadRequest)
		return
	}

	// Fetch owner (and recipient if provided)
	var users []models.User
	if requestBody.RecipientID != 0 {
		userIDs := []uint{userID, requestBody.RecipientID}
		if err := config.DB.Where("id IN ?", userIDs).Find(&users).Error; err != nil || len(users) != 2 {
			http.Error(w, "Users not found", http.StatusBadRequest)
			return
		}
	} else {
		if err := config.DB.Where("id = ?", userID).Find(&users).Error; err != nil || len(users) != 1 {
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}
	}

	// Create chatroom
	newChatRoom := models.Chatroom{
		OwnerId:      userID,
		Title:        requestBody.Title,
		IsPublic:     requestBody.IsPublic,
		MaxUserCount: requestBody.MaxUserCount,
	}
	if err := config.DB.Create(&newChatRoom).Error; err != nil {
		http.Error(w, "Failed to create chatroom", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	// Add entries to UserChatroom
	userChatrooms := []models.UserChatroom{
		{UserID: userID, Name: users[0].Name, ChatroomID: newChatRoom.Id, IsJoined: true, LastJoinTime: &now, IsAdmin: true},
	}
	if requestBody.RecipientID != 0 {
		userChatrooms = append(userChatrooms, models.UserChatroom{UserID: requestBody.RecipientID, Name: users[1].Name, ChatroomID: newChatRoom.Id, IsJoined: false, LastJoinTime: &now})
	}
	if err := config.DB.Create(&userChatrooms).Error; err != nil {
		http.Error(w, "Failed to link users to chatroom", http.StatusInternalServerError)
		return
	}

	// Respond with the created chatroom details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newChatRoom)
}

func DeleteChatroom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	isAdmin := r.Context().Value("isAdmin").(bool)

	if id == "" {
		http.Error(w, "Please provide a valid ID", http.StatusBadRequest)
		return
	}

	if !isAdmin {
		http.Error(w, "Unauthorized, user not an admin", http.StatusUnauthorized)
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

	// Delete associated UserChatroom entries
	if err := config.DB.Where("chatroom_id = ?", chatroom.Id).Delete(&models.UserChatroom{}).Error; err != nil {
		http.Error(w, "Error deleting user-chatroom links", http.StatusInternalServerError)
		return
	}

	// Delete chatroom
	if err := config.DB.Delete(&chatroom).Error; err != nil {
		http.Error(w, "Error deleting chatroom", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "Deleted chatroom successfully",
	})
}

func JoinChatroom(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized: missing or invalid user ID", http.StatusUnauthorized)
		return
	}
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized: missing or invalid username", http.StatusUnauthorized)
		return
	}

	chatroomID := r.PathValue("id")
	if chatroomID == "" {
		http.Error(w, "Please provide a valid chatroom ID", http.StatusBadRequest)
		return
	}

    var chatroom models.Chatroom
    if err := config.DB.First(&chatroom, chatroomID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            http.Error(w, "Chatroom not found", http.StatusNotFound)
        } else {
            http.Error(w, "Error retrieving chatroom", http.StatusInternalServerError)
        }
        return
    }
    // delegate to service to handle join/create membership (idempotent)
    if _, err := Svcs.Chat.JoinChatroom(userID, username, chatroomID); err != nil {
        http.Error(w, "Error adding user to chatroom", http.StatusInternalServerError)
        return
    }

    // reload membership for checks
    var userChatroom models.UserChatroom
    if err := config.DB.Where("user_id = ? AND chatroom_id = ?", userID, chatroom.Id).First(&userChatroom).Error; err == nil {
        if userChatroom.IsBanned {
            http.Error(w, "You are banned from this chatroom", http.StatusForbidden)
            return
        }
        if userChatroom.IsJoined {
            // already joined; return success idempotently
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]any{
                "Status":   "Already in chatroom",
                "Chatroom": chatroom,
            })
            return
        }
    }

    now := time.Now()
    if !chatroom.IsPublic {
        if userChatroom.IsInvited && userChatroom.InviteExpires.After(now) {
            userChatroom.IsInvited = false
        } else {
            http.Error(w, "You are not invited to this chatroom, or invitation has expired.", http.StatusForbidden)
            return
        }
    }

    userChatroom.IsJoined = true
    userChatroom.LastJoinTime = &now

    if err := config.DB.Save(&userChatroom).Error; err != nil {
        http.Error(w, "Error updating user-chatroom association", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "Status":   "User added to chatroom successfully",
        "Chatroom": chatroom,
    })

}

func transferOwnership(chatroomId string) error {
	var newOwner models.UserChatroom

	// Try to find an admin first
	err := config.DB.Where("chatroom_id = ? AND is_admin = ?", chatroomId, true).
		Order("last_join_time ASC").
		First(&newOwner).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No admin found, assign oldest non-admin user
		err = config.DB.Where("chatroom_id = ?", chatroomId).
			Order("last_join_time ASC").
			First(&newOwner).Error
	}

	// If no users are left, ownership transfer isn't needed
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	// Assign new owner
	newOwner.IsOwner = true
	if err := config.DB.Save(&newOwner).Error; err != nil {
		return err
	}

	// Update chatroom's owner ID
	return config.DB.Model(&models.Chatroom{}).
		Where("id = ?", chatroomId).
		Update("owner_id", newOwner.UserID).Error
}


func LeaveChatroom(w http.ResponseWriter, r *http.Request) {
    userId := r.Context().Value("userID").(uint)
    chatroomId := r.PathValue("id")

    utils.MembershipCache.Delete(fmt.Sprintf("membership:%v:%s", userId, chatroomId))
    wasOwner := false // best-effort; can enhance by loading membership
    res, err := Svcs.Chat.LeaveChatroom(userId, chatroomId, wasOwner)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    json.NewEncoder(w).Encode(res)
}

