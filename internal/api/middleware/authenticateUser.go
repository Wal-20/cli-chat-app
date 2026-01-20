package middleware

import (
	"context"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"net/http"
	"strings"
	"time"
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

		var claims map[string]any
		var err error

		// Validate (or get from cache) the access token only; no server file IO.
		if cachedClaims, found := utils.AuthCache.Get(tokenString); found {
			claims = cachedClaims.(map[string]any)
		} else {
			claims, err = utils.ValidateJWTToken(tokenString)
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			// Cache valid access token claims for a short period
			utils.AuthCache.Set(tokenString, claims, time.Minute*5)
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
