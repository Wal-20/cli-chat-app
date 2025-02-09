package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"os"	
	"fmt"
	"log"
	"encoding/json"
	"path/filepath"
)

type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

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


func GenerateJWTToken(userID uint) (string, error) {

	SECRET_KEY := os.Getenv("JWT_SECRET")
	if SECRET_KEY == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	claims := jwt.MapClaims{
		"userID": userID,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SECRET_KEY))

	return signedToken, err 
}

func GenerateRefreshToken(userID uint) (string, error) {
	SECRET_KEY := os.Getenv("JWT_SECRET")
	if SECRET_KEY == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	claims := jwt.MapClaims{
		"userID": userID,
		"exp": time.Now().Add(time.Hour * 168).Unix(), // 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(SECRET_KEY))

	return signedToken, err 
}

// ValidateJWTToken validates the JWT access token and returns the claims
func ValidateJWTToken(tokenString string) (jwt.MapClaims, error) {

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

