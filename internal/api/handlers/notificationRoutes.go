package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"net/http"
	"strconv"
)

func DeleteNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	idParam := r.PathValue("id")
	if idParam == "" {
		http.Error(w, `{"error": "missing notification ID"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid notification ID"}`, http.StatusBadRequest)
		return
	}

	result := config.DB.Delete(&models.Notification{}, id)
	if result.Error != nil {
		http.Error(w, `{"error": "failed to delete notification"}`, http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, `{"error": "notification not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder.Encode(map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("notification %d deleted successfully", id),
	})
}
