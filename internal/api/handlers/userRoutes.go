package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitializeDB(database *gorm.DB) {
	DB = database
}

	
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
		result := DB.First(&user, id)
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
		result := DB.Find(&users)

		if result.Error != nil {
			encoder.Encode(gin.H{"Message": "Failed to retrieve users"})
			return
		}

		encoder.Encode(gin.H{"Users": users})

	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	//encode the body into the user struct
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)

	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if user.Name == ""  || user.Password == "" {
		http.Error(w, "Invalid User", http.StatusBadRequest)
		return
	}

	result := DB.Create(&user)

	if result.Error != nil {
		http.Error(w, "Error creating user", http.StatusBadRequest)
		return
	}

	// Respond back with the user data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]interface{}{
		"Status": "User created successfully",
		"User":   user,
	})

}
