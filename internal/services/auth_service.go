package services

import (
    "errors"
    "time"
    "github.com/Wal-20/cli-chat-app/internal/models"
    "github.com/Wal-20/cli-chat-app/internal/repositories"
    "github.com/Wal-20/cli-chat-app/internal/utils"
    "gorm.io/gorm"
)

type AuthService struct {
    users repositories.UserRepository
}

func NewAuthService(users repositories.UserRepository) *AuthService { return &AuthService{users: users} }

func (s *AuthService) Login(username, password string) (access, refresh string, user models.User, err error) {
    u, err := s.users.FindByName(username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return "", "", models.User{}, gorm.ErrRecordNotFound
        }
        return "", "", models.User{}, err
    }
    if !utils.CheckPasswordHash(password, u.Password) {
        return "", "", models.User{}, errors.New("invalid password")
    }
    access, err = utils.GenerateJWTToken(u.ID, u.Name)
    if err != nil { return "", "", models.User{}, err }
    refresh, err = utils.GenerateRefreshToken(u.ID, u.Name)
    if err != nil { return "", "", models.User{}, err }
    now := time.Now()
    u.LastLogin = &now
    _ = s.users.Save(u)
    return access, refresh, *u, nil
}

func (s *AuthService) Register(username, password string) (access, refresh string, user models.User, err error) {
    hashed, err := utils.HashPassword(password)
    if err != nil { return "", "", models.User{}, err }
    u := models.User{Name: username, Password: hashed}
    now := time.Now()
    u.LastLogin = &now
    if err := s.users.Create(&u); err != nil { return "", "", models.User{}, err }
    access, err = utils.GenerateJWTToken(u.ID, u.Name)
    if err != nil { return "", "", models.User{}, err }
    refresh, err = utils.GenerateRefreshToken(u.ID, u.Name)
    if err != nil { return "", "", models.User{}, err }
    return access, refresh, u, nil
}

func (s *AuthService) GetChatroomsByUser(userID uint) ([]models.Chatroom, error) {
    return s.users.GetChatroomsByUserID(userID)
}

