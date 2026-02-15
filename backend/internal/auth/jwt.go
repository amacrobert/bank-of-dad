package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenTTL = 15 * time.Minute

type contextKey string

const (
	ContextKeyUserType contextKey = "user_type"
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyFamilyID contextKey = "family_id"
)

func GetUserType(r *http.Request) string {
	val, _ := r.Context().Value(ContextKeyUserType).(string)
	return val
}

func GetUserID(r *http.Request) int64 {
	val, _ := r.Context().Value(ContextKeyUserID).(int64)
	return val
}

func GetFamilyID(r *http.Request) int64 {
	val, _ := r.Context().Value(ContextKeyFamilyID).(int64)
	return val
}

type Claims struct {
	UserType string `json:"user_type"`
	UserID   int64  `json:"user_id"`
	FamilyID int64  `json:"family_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed HS256 JWT with the given user info.
// The token expires after 15 minutes.
func GenerateAccessToken(jwtKey []byte, userType string, userID int64, familyID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserType: userType,
		UserID:   userID,
		FamilyID: familyID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%s:%d", userType, userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateAccessToken parses and validates a JWT string.
// It enforces HS256 to prevent algorithm confusion attacks.
func ValidateAccessToken(jwtKey []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
