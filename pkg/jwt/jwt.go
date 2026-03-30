package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service defines JWT operations
type Service interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	RefreshAccessToken(refreshToken string) (string, error)
}

// service implements Service
type service struct {
	secret        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewService creates a new JWT service
func NewService(secret string, accessExpiry, refreshExpiry time.Duration) Service {
	return &service{
		secret:        secret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// Claims is a struct for JWT claims that include standard claims and custom fields.
type Claims struct {
	jwt.RegisteredClaims
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
}

// GenerateAccessToken generates a new JWT access token for the given user ID and roles.
func GenerateAccessToken(userID string, roles []string, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Roles:  roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken generates a random refresh token string.
func GenerateRefreshToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateToken validates and parses a JWT token string using typed Claims.
// It enforces standard validations like expiry. Returns typed Claims on success.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ExtractUserID extracts the user ID from JWT claims.
func ExtractUserID(claims jwt.MapClaims) (primitive.ObjectID, error) {
	userID, ok := claims["user_id"].(string)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("user_id not found in claims")
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid user_id format: %w", err)
	}

	return objID, nil
}

// ExtractRoles extracts roles from either a typed Claims or a MapClaims payload.
// It first attempts to read []string, and falls back to []interface{} conversion.
func ExtractRoles(claims jwt.MapClaims) ([]string, error) {
	raw, ok := claims["roles"]
	if !ok {
		return nil, fmt.Errorf("roles not found in claims")
	}

	// Try []string directly
	if rs, ok := raw.([]string); ok {
		return rs, nil
	}

	// Fallback to []interface{}
	if rs, ok := raw.([]interface{}); ok {
		out := make([]string, len(rs))
		for i, r := range rs {
			s, ok := r.(string)
			if !ok {
				return nil, fmt.Errorf("invalid role at index %d", i)
			}
			out[i] = s
		}
		return out, nil
	}

	return nil, fmt.Errorf("unsupported roles format in claims")
}

// GenerateAccessToken generates a new JWT access token for the given user ID and roles.
func (s *service) GenerateAccessToken(userID string, roles []string) (string, error) {
	return GenerateAccessToken(userID, roles, s.secret, s.accessExpiry)
}

// GenerateRefreshToken generates a new JWT refresh token for the given user ID.
func (s *service) GenerateRefreshToken(userID string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Roles:  []string{}, // Refresh tokens don't need roles
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ValidateAccessToken validates and parses a JWT access token string.
func (s *service) ValidateAccessToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString, s.secret)
}

// ValidateRefreshToken validates and parses a JWT refresh token string.
func (s *service) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString, s.secret)
}

// RefreshAccessToken generates a new access token using a valid refresh token.
func (s *service) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	return s.GenerateAccessToken(claims.UserID, claims.Roles)
}
