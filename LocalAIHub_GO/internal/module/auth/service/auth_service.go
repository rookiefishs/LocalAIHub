package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"localaihub/localaihub_go/internal/module/auth/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)

type Session struct {
	Token    string
	AdminID  int64
	Username string
}

type AuthService struct {
	repo      *repository.AdminRepository
	jwtSecret string
}

func NewAuthService(repo *repository.AdminRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password, ip, userAgent string) (*Session, error) {
	admin, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil || admin.Status != "active" {
		_ = s.repo.CreateLoginLog(ctx, nil, username, "login_failed", ip, userAgent, "admin not found or disabled")
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		adminID := admin.ID
		_ = s.repo.CreateLoginLog(ctx, &adminID, username, "login_failed", ip, userAgent, "password mismatch")
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(admin.ID, admin.Username)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateLastLogin(ctx, admin.ID, ip)
	adminID := admin.ID
	_ = s.repo.CreateLoginLog(ctx, &adminID, username, "login_success", ip, userAgent, "ok")

	return &Session{Token: token, AdminID: admin.ID, Username: admin.Username}, nil
}

func (s *AuthService) generateToken(adminID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"admin_id":  adminID,
		"username":  username,
		"exp":       time.Now().Add(7 * 24 * time.Hour).Unix(),
		"issued_at": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) Authenticate(tokenString string) (*Session, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	adminID, ok := claims["admin_id"].(float64)
	if !ok {
		return nil, ErrUnauthorized
	}
	username, _ := claims["username"].(string)

	return &Session{
		Token:    tokenString,
		AdminID:  int64(adminID),
		Username: username,
	}, nil
}

func (s *AuthService) CurrentAdmin(ctx context.Context, adminID int64) (*repository.Admin, error) {
	return s.repo.GetByID(ctx, adminID)
}
