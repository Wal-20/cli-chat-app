package middleware

import (
	"fmt"
	"net/http"
	"strings"
    "context"
    "os"
	"github.com/golang-jwt/jwt/v5"
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
		claims, err := validateAccessToken(tokenString)
		if err != nil {
			// Check for a refresh token if the access token is invalid
			tokenPair, err := utils.LoadTokenPair() // Load the refresh token from the user's device
			if err != nil || tokenPair.RefreshToken == "" {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Validate the refresh token and issue a new access token
			refreshClaims, err := validateAccessToken(tokenPair.RefreshToken)
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Extract userID from the refresh token
			userIDFloat, ok := refreshClaims["userID"].(float64)
			if !ok {
				http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
				return
			}
			userID := uint(userIDFloat)

			// Generate a new access token
			newAccessToken, err := utils.GenerateJWTToken(userID)
			if err != nil {
				http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
				return
			}

			tokenPair.AccessToken = newAccessToken
			utils.SaveTokenPair(tokenPair)

			// Send the new access token back to the client
			http.SetCookie(w, &http.Cookie{
				Name:  "access_token",
				Value: newAccessToken,
				Path:  "/",
				// Optional: Set expiry for the cookie if needed
			})

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

// ValidateAccessToken validates the JWT access token and returns the claims
func validateAccessToken(tokenString string) (jwt.MapClaims, error) {

	SECRET_KEY := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is valid
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET_KEY, nil
	})

	if err != nil {
		return nil, err
	}

	// Return the claims if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

