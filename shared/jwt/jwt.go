// Package jwt provides JWT token generation and validation.
// Used by auth-service (generate) and all other services via middleware (validate).
//
// USAGE:
//   // Generate (auth-service only)
//   tokens, err := jwt.GenerateTokenPair(jwt.Claims{UserID: "123", Role: "user"}, secret)
//
//   // Validate (all services via middleware)
//   claims, err := jwt.ValidateToken(tokenString, secret)

package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims holds the data embedded in each JWT token.
// Keep this small — it's in every request.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`   // "user" | "admin" | "agent"
	jwtlib.RegisteredClaims
}

// TokenPair contains both access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp
}

// GenerateTokenPair creates a short-lived access token + long-lived refresh token.
//
// Access token:  15 minutes — used for API calls
// Refresh token: 7 days    — used to get a new access token
func GenerateTokenPair(userID, email, role, secret string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(15 * time.Minute)

	// ── Access Token ──
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(accessExpiry),
			IssuedAt:  jwtlib.NewNumericDate(now),
			ID:        uuid.New().String(), // unique token ID (jti)
		},
	}
	accessToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	// ── Refresh Token ──
	refreshClaims := jwtlib.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwtlib.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		IssuedAt:  jwtlib.NewNumericDate(now),
		ID:        uuid.New().String(),
	}
	refreshToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}

// ValidateToken parses and validates a JWT string.
// Returns the claims if valid, error if expired or tampered.
func ValidateToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwtlib.Token) (interface{}, error) {
			// Ensure the signing method is what we expect
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
