package auth

import (
	"context"
	"errors"
	"time"

	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	queries   *domain.Queries
	jwtSecret string
}

func NewAuthService(database *db.DB, secret string) *AuthService {
	return &AuthService{
		queries:   domain.New(database.Pool),
		jwtSecret: secret,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type Claims struct {
	UserID   string   `json:"userId"`
	TenantID string   `json:"tenantId"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	// 1. Fetch user by email
	user, err := s.queries.GetUserForLogin(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials") // Prevent user enumeration
	}

	// 2. Verify hashed password
	if !user.PasswordHash.Valid || !s.CheckPassword(password, user.PasswordHash.String) {
		return "", errors.New("invalid credentials")
	}

	// 3. Fetch User Roles
	roles, err := s.queries.GetUserRoles(ctx, domain.GetUserRolesParams{
		TenantID: user.TenantID,
		UserID:   user.ID,
	})
	if err != nil {
		roles = []string{} // Default to empty array on failure
	}

	// 4. Generate JWT Token
	expirationTime := time.Now().Add(24 * time.Hour)

	// Convert UUID bytes directly to 36 char string format
	parsedUserID, _ := uuid.FromBytes(user.ID.Bytes[:])
	userIDStr := parsedUserID.String()

	parsedTenantID, _ := uuid.FromBytes(user.TenantID.Bytes[:])
	tenantIDStr := parsedTenantID.String()

	claims := &Claims{
		UserID:   userIDStr,
		TenantID: tenantIDStr,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
