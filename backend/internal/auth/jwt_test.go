package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKey = []byte("test-secret-key-that-is-at-least-32-bytes-long!!")

func TestGenerateAndValidateAccessToken(t *testing.T) {
	tokenStr, err := GenerateAccessToken(testKey, "parent", 123, 42)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	claims, err := ValidateAccessToken(testKey, tokenStr)
	require.NoError(t, err)
	assert.Equal(t, "parent", claims.UserType)
	assert.Equal(t, int64(123), claims.UserID)
	assert.Equal(t, int64(42), claims.FamilyID)
	assert.Equal(t, "parent:123", claims.Subject)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	now := time.Now().Add(-1 * time.Hour)
	claims := Claims{
		UserType: "child",
		UserID:   456,
		FamilyID: 10,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "child:456",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)), // expired 45 min ago
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(testKey)
	require.NoError(t, err)

	_, err = ValidateAccessToken(testKey, tokenStr)
	assert.Error(t, err)
}

func TestValidateAccessToken_Tampered(t *testing.T) {
	tokenStr, err := GenerateAccessToken(testKey, "parent", 1, 1)
	require.NoError(t, err)

	// Tamper with the token by changing a character
	tampered := tokenStr[:len(tokenStr)-2] + "XX"
	_, err = ValidateAccessToken(testKey, tampered)
	assert.Error(t, err)
}

func TestValidateAccessToken_WrongKey(t *testing.T) {
	tokenStr, err := GenerateAccessToken(testKey, "parent", 1, 1)
	require.NoError(t, err)

	wrongKey := []byte("wrong-key-that-is-also-at-least-32-bytes-long!!")
	_, err = ValidateAccessToken(wrongKey, tokenStr)
	assert.Error(t, err)
}

func TestValidateAccessToken_WrongSigningMethod(t *testing.T) {
	// Create a token signed with a different method (none)
	claims := Claims{
		UserType: "parent",
		UserID:   1,
		FamilyID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "parent:1",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = ValidateAccessToken(testKey, tokenStr)
	assert.Error(t, err)
}

func TestGenerateAccessToken_ChildSubject(t *testing.T) {
	tokenStr, err := GenerateAccessToken(testKey, "child", 456, 10)
	require.NoError(t, err)

	claims, err := ValidateAccessToken(testKey, tokenStr)
	require.NoError(t, err)
	assert.Equal(t, "child:456", claims.Subject)
	assert.Equal(t, "child", claims.UserType)
	assert.Equal(t, int64(456), claims.UserID)
	assert.Equal(t, int64(10), claims.FamilyID)
}
