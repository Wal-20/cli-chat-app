package middleware

import (
	"net/http"
	"strings"
	"time"
    "context"
	"github.com/Wal-20/cli-chat-app/internal/utils"
)


func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		var claims map[string]interface{}
		var err error

		// Check access token cache
		if cachedClaims, found := utils.AuthCache.Get(tokenString); found {
			claims = cachedClaims.(map[string]interface{})
		} else {
			claims, err = utils.ValidateJWTToken(tokenString)
			if err != nil {
				// Access token is invalid, try refresh token
				tokenPair, err := utils.LoadTokenPair()
				if err != nil || tokenPair.RefreshToken == "" {
					http.Error(w, "Authentication required", http.StatusUnauthorized)
					return
				}

				refreshToken := tokenPair.RefreshToken

				// Try refresh token from cache
				var refreshClaims map[string]interface{}
				if cachedRefreshClaims, found := utils.AuthCache.Get(refreshToken); found {
					refreshClaims = cachedRefreshClaims.(map[string]interface{})
				} else {
					refreshClaims, err = utils.ValidateJWTToken(refreshToken)
					if err != nil {
						http.Error(w, "Authentication required", http.StatusUnauthorized)
						return
					}
				}

				// Extract and validate required fields
				username, exists := refreshClaims["username"]
				userIDValue, idExists := refreshClaims["userID"]
				if !exists || !idExists {
					http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
					return
				}

				userID, ok := userIDValue.(float64)
				if !ok {
					http.Error(w, "Invalid userID format", http.StatusUnauthorized)
					return
				}

				// Generate and cache new access token
				newAccessToken, err := utils.GenerateJWTToken(uint(userID), username.(string))
				if err != nil {
					http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
					return
				}
				tokenPair.AccessToken = newAccessToken
				utils.SaveTokenPair(tokenPair)

				// Cache new token
				claims, _ = utils.ValidateJWTToken(newAccessToken)
				utils.AuthCache.Set(newAccessToken, claims, time.Minute * 5)

				// Add user context and proceed
				ctx := context.WithValue(r.Context(), "userID", uint(userID))
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
			// Cache valid access token claims
			utils.AuthCache.Set(tokenString, claims, time.Minute * 5)
		}

		// Extract user info from claims
		userIDFloat, ok := claims["userID"].(float64)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		username, ok := claims["username"]
		if !ok {
			http.Error(w, "Missing username in token claims", http.StatusUnauthorized)
			return
		}

		// Add user context
		ctx := context.WithValue(r.Context(), "userID", uint(userIDFloat))
		ctx = context.WithValue(ctx, "username", username)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}



