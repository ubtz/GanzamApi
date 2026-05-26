package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"GanzamApi/models"
	"GanzamApi/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type AuthService struct {
	store repositories.UserStore
}

func NewAuthService(store repositories.UserStore) *AuthService {
	return &AuthService{store: store}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, string, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if req.Phone == "" || req.Password == "" {
		return nil, "", errors.New("phone and password are required")
	}

	if _, err := s.store.FindByPhone(ctx, req.Phone); err == nil {
		return nil, "", ErrUserExists
	} else if !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user, err := s.store.Create(ctx, req, string(passwordHash))
	if err != nil {
		return nil, "", err
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.User, string, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" || req.Password == "" {
		return nil, "", errors.New("phone and password are required")
	}

	user, err := s.store.FindByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func generateToken(user *models.User) (string, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(map[string]interface{}{
		"sub":   fmt.Sprintf("%d", user.ID),
		"phone": user.Phone,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsigned := header + "." + payload

	mac := hmac.New(sha256.New, []byte(getJWTSecret()))
	_, err = mac.Write([]byte(unsigned))
	if err != nil {
		return "", err
	}

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsigned + "." + signature, nil
}

func getJWTSecret() string {
	if value := strings.TrimSpace(os.Getenv("JWT_SECRET")); value != "" {
		return value
	}

	return "ganzam-secret"
}
