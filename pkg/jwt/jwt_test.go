package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/pkg/jwt"
)

func TestCreateJWT(t *testing.T) {
	signingKey := []byte("test-signing-key")
	username := "testuser"

	t.Run("Valid token creation", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		token, err := jwt.CreateJWT(username, signingKey, expiresAt)
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("Expired token should still be created", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		token, err := jwt.CreateJWT(username, signingKey, expiresAt)
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})
}

func TestParseJWT(t *testing.T) {
	signingKey := []byte("test-signing-key")
	username := "testuser"
	expiresAt := time.Now().Add(1 * time.Hour)

	validToken, err := jwt.CreateJWT(username, signingKey, expiresAt)
	require.NoError(t, err)

	t.Run("Valid token parsing", func(t *testing.T) {
		claims, err := jwt.ParseJWT(validToken, signingKey)
		require.NoError(t, err)
		require.NotNil(t, claims)
		require.Equal(t, username, claims.Subject)
	})

	t.Run("Invalid signing key", func(t *testing.T) {
		wrongKey := []byte("wrong-key")
		claims, err := jwt.ParseJWT(validToken, wrongKey)
		require.Error(t, err)
		require.Nil(t, claims)
	})

	t.Run("Invalid token format", func(t *testing.T) {
		claims, err := jwt.ParseJWT("invalid-token", signingKey)
		require.Error(t, err)
		require.Nil(t, claims)
	})

	t.Run("Expired token parsing", func(t *testing.T) {
		expiredToken, err := jwt.CreateJWT(username, signingKey, time.Now().Add(-1*time.Hour))
		require.NoError(t, err)
		claims, err := jwt.ParseJWT(expiredToken, signingKey)
		require.Error(t, err)
		require.Nil(t, claims)
	})
}
