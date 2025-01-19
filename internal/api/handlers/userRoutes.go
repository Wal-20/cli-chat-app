package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func GetUsers(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil || id <= 0 {

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(gin.H{"Status": "invalid ID"})
			return
		}

		var user models.User
		result := config.DB.First(&user, id)
		if result.Error != nil {

			w.WriteHeader(http.StatusNotFound)
			encoder.Encode(gin.H{"Status": "User not found"})
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder.Encode(gin.H{
			"User": user,
		})

	} else {
		var users []models.User
		result := config.DB.Find(&users)

		if result.Error != nil {
			encoder.Encode(gin.H{"Message": "Failed to retrieve users"})
			return
		}

		encoder.Encode(gin.H{"Users": users})

	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var userDB models.User
	result := config.DB.Where("name = ?", user.Name).First(&userDB)

	if result.Error != nil {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	if !utils.CheckPasswordHash(user.Password, userDB.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Load token pair from storage
	tokenPair, err := utils.LoadTokenPair()
	if err == nil && tokenPair.AccessToken != "" {
		// Validate the existing access token
		_, err := utils.ValidateJWTToken(tokenPair.AccessToken)
		if err == nil {
			// Token is valid; no need to generate a new one
			encoder.Encode(map[string]interface{}{
				"Status":      "user already logged in",
				"AccessToken": tokenPair.AccessToken,
			})
			return
		}
	}

	// Generate both access and refresh tokens
	accessToken, err := utils.GenerateJWTToken(userDB.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(userDB.ID)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Save tokens to local tokenPair
	tokenPair = utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving credentials", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	userDB.LastLogin = &now

	if err := config.DB.Save(&userDB).Error; err != nil {
		log.Printf("Failed to update LastLogin: %v", err)
		http.Error(w, "Failed to update LastLogin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]interface{}{
		"Status":       "Login successful",
		"Token":        accessToken,
		"RefreshToken": refreshToken,
	})

}


func LogOut(w http.ResponseWriter, r *http.Request) {
	// Delete the refresh token from the user's device
	
	tokenPair, err := utils.LoadTokenPair()
	if err != nil {
		http.Error(w, "Error loading token pair", http.StatusInternalServerError)
		return
	}

	tokenPair.RefreshToken = ""
	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving token pair", http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)

	encoder.Encode(map[string]interface{}{
		"Status": "Logged out succcessfully",
	})
}


func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if user.Name == "" || user.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusBadRequest)
		return
	}

	user.Password = hashedPassword

	now := time.Now()
	user.LastLogin = &now

	result := config.DB.Create(&user)

	if result.Error != nil {
		http.Error(w, "Error creating user", http.StatusBadRequest)
		return
	}

	// Generate both access and refresh tokens
	accessToken, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Save tokens to local tokenPair
	tokenPair := utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if err := utils.SaveTokenPair(tokenPair); err != nil {
		http.Error(w, "Error saving credentials", http.StatusInternalServerError)
		return
	}

	// Respond back with the user data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]interface{}{
		"Status": "User created successfully",
		"User":   user,
		"Access Token": accessToken,
		"Refresh Token": refreshToken,
	})

}


func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: add admin authorization for updating a user, now only the user can update themselves
	userID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(w,"Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating user", http.StatusInternalServerError)
		}
		return
	}

	var userUpdate models.User
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&userUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	
	if userUpdate.Name == "" || userUpdate.Password == "" {
		http.Error(w, "Missing attributes", http.StatusBadRequest)
		return
	}
	user.Name = userUpdate.Name

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	config.DB.Save(&user)

	encoder.Encode(map[string]interface{} {
		"Status": "User Updated successfully",
		"User": user,
	})
}

// todo: add delete user route, make sure that user's JWT token gets removed as well
