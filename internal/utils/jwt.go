package utils

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
	"path/filepath"
	"time"
)

type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// DefaultJWTSecret can be injected at build time (via -ldflags)
// for clients that don't have a runtime JWT_SECRET environment variable.
var DefaultJWTSecret string

func GetTokenPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cli-chat-config.json")
}

func SaveTokenPair(tokenPair TokenPair) error {
	data, err := json.Marshal(tokenPair)
	if err != nil {
		return err
	}
	return os.WriteFile(GetTokenPath(), data, 0600)
}

func LoadTokenPair() (TokenPair, error) {
	var tokenPair TokenPair
	data, err := os.ReadFile(GetTokenPath())
	if err != nil {
		return tokenPair, err
	}
	err = json.Unmarshal(data, &tokenPair)
	return tokenPair, err
}

func getJWTSecret() (string, error) {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret, nil
	}
	if DefaultJWTSecret != "" {
		return DefaultJWTSecret, nil
	}
	return "", fmt.Errorf("JWT_SECRET is not set")
}

func GenerateJWTToken(userID uint, username string) (string, error) {
	SECRET_KEY, err := getJWTSecret()
	if err != nil {
		log.Fatal(err)
	}
	claims := jwt.MapClaims{
		"userID":   userID,
		"username": username,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SECRET_KEY))

	return signedToken, err
}

func GenerateRefreshToken(userID uint, username string) (string, error) {
	SECRET_KEY, err := getJWTSecret()
	if err != nil {
		log.Fatal(err)
	}

	claims := jwt.MapClaims{
		"userID":   userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 168).Unix(), // 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(SECRET_KEY))

	return signedToken, err
}

// ValidateJWTToken validates the JWT access token and returns the claims
func ValidateJWTToken(tokenString string) (jwt.MapClaims, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}
	SECRET_KEY := []byte(secret)
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
		// Check for token expiration
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0) // Convert expiration to time.Time
			if expirationTime.Before(time.Now()) {
				return nil, fmt.Errorf("token has expired")
			}
		}
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func GetClaimsFromToken(tokenString string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token expired")
			}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
