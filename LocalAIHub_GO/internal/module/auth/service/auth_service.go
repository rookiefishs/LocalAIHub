package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"localaihub/localaihub_go/internal/module/auth/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrRegistrationClosed = errors.New("registration is closed")
)

type Session struct {
	Token        string
	RefreshToken string
	AdminID      int64
	Username     string
}

type AuthService struct {
	repo                *repository.AdminRepository
	jwtSecret           string
	registrationEnabled bool
	audit               interface {
		Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
	}
}

func NewAuthService(repo *repository.AdminRepository, jwtSecret string, registrationEnabled bool, audit interface {
	Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
}) *AuthService {
	return &AuthService{
		repo:                repo,
		jwtSecret:           jwtSecret,
		registrationEnabled: registrationEnabled,
		audit:               audit,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password, ip, userAgent string) (*Session, error) {
	admin, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil || admin.Status != "active" {
		if err := s.repo.CreateLoginLog(ctx, nil, username, "login_failed", ip, userAgent, "admin not found or disabled"); err != nil {
			logger.Log.Error().Err(err).Str("username", username).Msg("failed to create failed login log")
		}
		if s.audit != nil {
			s.audit.Log(ctx, "login", "admin_user", nil, map[string]any{"username": username, "result": "failed", "reason": "admin not found or disabled", "ip": ip}, ip, userAgent)
		}
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		adminID := admin.ID
		if err := s.repo.CreateLoginLog(ctx, &adminID, username, "login_failed", ip, userAgent, "password mismatch"); err != nil {
			logger.Log.Error().Err(err).Int64("admin_id", adminID).Msg("failed to create password mismatch login log")
		}
		if s.audit != nil {
			s.audit.Log(ctx, "login", "admin_user", &adminID, map[string]any{"username": username, "result": "failed", "reason": "password mismatch", "ip": ip}, ip, userAgent)
		}
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(admin.ID, admin.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(admin.ID, admin.Username)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateLastLogin(ctx, admin.ID, ip); err != nil {
		logger.Log.Error().Err(err).Int64("admin_id", admin.ID).Msg("failed to update last login")
	}
	adminID := admin.ID
	if err := s.repo.CreateLoginLog(ctx, &adminID, username, "login_success", ip, userAgent, "ok"); err != nil {
		logger.Log.Error().Err(err).Int64("admin_id", adminID).Msg("failed to create success login log")
	}
	if s.audit != nil {
		s.audit.Log(ctx, "login", "admin_user", &adminID, map[string]any{"username": admin.Username, "result": "success", "ip": ip}, ip, userAgent)
	}

	return &Session{Token: token, RefreshToken: refreshToken, AdminID: admin.ID, Username: admin.Username}, nil
}

func (s *AuthService) generateToken(adminID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"admin_id":  adminID,
		"username":  username,
		"type":      "access",
		"exp":       time.Now().Add(1 * time.Hour).Unix(),
		"issued_at": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) generateRefreshToken(adminID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"admin_id":  adminID,
		"username":  username,
		"type":      "refresh",
		"exp":       time.Now().Add(7 * 24 * time.Hour).Unix(),
		"issued_at": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) RefreshToken(refreshTokenString string) (*Session, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (any, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, ErrUnauthorized
	}

	adminID, ok := claims["admin_id"].(float64)
	if !ok {
		return nil, ErrUnauthorized
	}
	username, _ := claims["username"].(string)

	newToken, err := s.generateToken(int64(adminID), username)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(int64(adminID), username)
	if err != nil {
		return nil, err
	}

	return &Session{Token: newToken, RefreshToken: newRefreshToken, AdminID: int64(adminID), Username: username}, nil
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
	if username == "" {
		admin, lookupErr := s.repo.GetByID(context.Background(), int64(adminID))
		if lookupErr == nil && admin != nil {
			username = admin.Username
		}
	}
	if username == "" {
		username = fmt.Sprintf("admin#%d", int64(adminID))
	}

	return &Session{
		Token:    tokenString,
		AdminID:  int64(adminID),
		Username: username,
	}, nil
}

func (s *AuthService) CurrentAdmin(ctx context.Context, adminID int64) (*repository.Admin, error) {
	return s.repo.GetByID(ctx, adminID)
}

func (s *AuthService) Register(ctx context.Context, username, password, ip, userAgent string) (*Session, error) {
	if !s.registrationEnabled {
		return nil, ErrRegistrationClosed
	}

	if len(username) < 3 || len(username) > 64 {
		return nil, fmt.Errorf("username must be between 3 and 64 characters")
	}

	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	existing, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	adminID, err := s.repo.CreateAdmin(ctx, username, string(passwordHash))
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}

	if s.audit != nil {
		id := adminID
		s.audit.Log(ctx, "register", "admin_user", &id, map[string]any{"username": username, "ip": ip}, ip, userAgent)
	}

	token, err := s.generateToken(adminID, username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(adminID, username)
	if err != nil {
		return nil, err
	}

	return &Session{Token: token, RefreshToken: refreshToken, AdminID: adminID, Username: username}, nil
}
