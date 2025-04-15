package middleware

import (
	"net/http"
	"strings"
    "context"
	"github.com/Wal-20/cli-chat-app/internal/utils"
)

// authenticateing the user
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the authorization header is present
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract the token from the Authorization header (Bearer <token>)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Validate the access token
		claims, err := utils.ValidateJWTToken(tokenString)
		if err != nil {
			// Check for a refresh token if the access token is invalid
			tokenPair, err := utils.LoadTokenPair() // Load the refresh token from the user's device
			if err != nil || tokenPair.RefreshToken == "" {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Validate the refresh token and issue a new access token
			refreshClaims, err := utils.ValidateJWTToken(tokenPair.RefreshToken)
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			username, exists := refreshClaims["username"]
			if !exists || username == nil {
				http.Error(w, "Invalid refresh token: missing username", http.StatusUnauthorized)
				return
			}

			userIDValue, exists := refreshClaims["userID"]
			if !exists || userIDValue == nil {
				http.Error(w, "Invalid refresh token: missing user ID", http.StatusUnauthorized)
				return
			}

			userIDFloat, ok := userIDValue.(float64)
			if !ok {
				http.Error(w, "Invalid refresh token: incorrect user ID format", http.StatusUnauthorized)
				return
			}

			userID := uint(userIDFloat)

			// Generate a new access token
			newAccessToken, err := utils.GenerateJWTToken(userID, username.(string))
			if err != nil {
				http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
				return
			}

			tokenPair.AccessToken = newAccessToken
			utils.SaveTokenPair(tokenPair)

			ctx := context.WithValue(r.Context(), "userID", userID)
			r = r.WithContext(ctx)

			// Proceed with the next handler
			next.ServeHTTP(w, r)
			return
		}

		// Extract userID from the valid access token
		userIDFloat, ok := claims["userID"].(float64)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		userID := uint(userIDFloat)

		// Store the userID in the context
		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}


