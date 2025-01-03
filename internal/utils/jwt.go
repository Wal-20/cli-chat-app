package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"os"	
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

