package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

func TestFileAuthRepository(t *testing.T) {
	ctx := context.Background()

	tmpFile := filepath.Join(t.TempDir(), "test_auth.json")

	t.Run("Create and Get User", func(t *testing.T) {
		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		user := domain.User{
			Login:    "testuser",
			Password: "testpass",
		}

		err = repo.CreateUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		assert.Equal(t, "testuser", retrieved.Login)
		assert.Equal(t, "testpass", retrieved.Password)
		assert.NotEmpty(t, retrieved.ID)

		retrievedByID, err := repo.GetUserByID(ctx, retrieved.ID)
		require.NoError(t, err)
		assert.Equal(t, retrieved.ID, retrievedByID.ID)
	})

	t.Run("Duplicate User Error", func(t *testing.T) {
		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		user := domain.User{Login: "duplicate", Password: "pass1"}
		err = repo.CreateUser(ctx, user)
		require.NoError(t, err)

		user2 := domain.User{Login: "duplicate", Password: "pass2"}
		err = repo.CreateUser(ctx, user2)
		assert.ErrorIs(t, err, appErrors.ErrAlreadyRegistered)
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		_, err = repo.GetUserByLogin(ctx, "nonexistent")
		assert.ErrorIs(t, err, appErrors.ErrUserNotFound)

		_, err = repo.GetUserByID(ctx, "nonexistent-id")
		assert.ErrorIs(t, err, appErrors.ErrUserNotFound)
	})

	t.Run("Auto Generate ID", func(t *testing.T) {
		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		user := domain.User{Login: "autoid", Password: "pass"}
		err = repo.CreateUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByLogin(ctx, "autoid")
		require.NoError(t, err)
		assert.NotEmpty(t, retrieved.ID)
	})

	t.Run("Data Persistence", func(t *testing.T) {
		repo1, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		user := domain.User{Login: "persistent", Password: "pass"}
		err = repo1.CreateUser(ctx, user)
		require.NoError(t, err)

		repo2, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		retrieved, err := repo2.GetUserByLogin(ctx, "persistent")
		require.NoError(t, err)
		assert.Equal(t, "persistent", retrieved.Login)
	})

	t.Run("Multiple Users", func(t *testing.T) {
		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		users := []string{"user1", "user2", "user3"}
		for _, login := range users {
			err := repo.CreateUser(ctx, domain.User{Login: login, Password: "pass"})
			require.NoError(t, err)
		}

		for _, login := range users {
			_, err := repo.GetUserByLogin(ctx, login)
			assert.NoError(t, err)
		}
	})
}

func TestFileAuthRepository_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Corrupted JSON File", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "corrupted.json")

		err := os.WriteFile(tmpFile, []byte("{invalid json}"), 0644)
		require.NoError(t, err)

		_, err = NewFileAuthRepository(tmpFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("Empty File", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "empty.json")

		err := os.WriteFile(tmpFile, []byte{}, 0644)
		require.NoError(t, err)

		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		_, err = repo.GetUserByLogin(ctx, "anyuser")
		assert.Error(t, err)
	})

	t.Run("Load Existing Data", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "existing.json")

		existingData := `{
			"existing_user": {
				"id": "user-123",
				"login": "existing_user", 
				"password": "hashed_pass"
			}
		}`
		err := os.WriteFile(tmpFile, []byte(existingData), 0644)
		require.NoError(t, err)

		repo, err := NewFileAuthRepository(tmpFile)
		require.NoError(t, err)

		user, err := repo.GetUserByLogin(ctx, "existing_user")
		require.NoError(t, err)
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, "existing_user", user.Login)
	})
}
