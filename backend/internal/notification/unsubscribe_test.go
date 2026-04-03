package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateUnsubscribeToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!")

	token, err := GenerateUnsubscribeToken(42, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Valid token round-trips
	parentID, err := ValidateUnsubscribeToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), parentID)
}

func TestValidateUnsubscribeToken_TamperedToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!")

	token, err := GenerateUnsubscribeToken(42, secret)
	require.NoError(t, err)

	// Tamper with the token
	tampered := token[:len(token)-2] + "XX"
	_, err = ValidateUnsubscribeToken(tampered, secret)
	assert.Error(t, err)
}

func TestValidateUnsubscribeToken_WrongSecret(t *testing.T) {
	secret1 := []byte("test-secret-key-at-least-32-bytes!")
	secret2 := []byte("different-secret-key-at-least-32b!")

	token, err := GenerateUnsubscribeToken(42, secret1)
	require.NoError(t, err)

	_, err = ValidateUnsubscribeToken(token, secret2)
	assert.Error(t, err)
}

func TestValidateUnsubscribeToken_InvalidFormat(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!")

	_, err := ValidateUnsubscribeToken("not-a-valid-token", secret)
	assert.Error(t, err)

	_, err = ValidateUnsubscribeToken("", secret)
	assert.Error(t, err)
}

func TestGenerateUnsubscribeToken_ContainsCorrectParentID(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!")

	for _, id := range []int64{1, 100, 999999} {
		token, err := GenerateUnsubscribeToken(id, secret)
		require.NoError(t, err)

		parentID, err := ValidateUnsubscribeToken(token, secret)
		require.NoError(t, err)
		assert.Equal(t, id, parentID)
	}
}
