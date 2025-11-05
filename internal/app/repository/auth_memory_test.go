package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

func TestMemoryAuthRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and Get User", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		user := domain.User{
			Login:    "testuser",
			Password: "testpass",
			Token:    "token123",
		}

		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByLogin(ctx, "testuser")
		require.NoError(t, err)
		assert.Equal(t, "testuser", retrieved.Login)
		assert.Equal(t, "testpass", retrieved.Password)
		assert.Equal(t, "token123", retrieved.Token)
		assert.NotEmpty(t, retrieved.ID)

		retrievedByID, err := repo.GetUserByID(ctx, retrieved.ID)
		require.NoError(t, err)
		assert.Equal(t, retrieved.ID, retrievedByID.ID)
	})

	t.Run("Duplicate User Error", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		user := domain.User{Login: "duplicate", Password: "pass1"}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		user2 := domain.User{Login: "duplicate", Password: "pass2"}
		err = repo.CreateUser(ctx, user2)
		assert.ErrorIs(t, err, appErrors.ErrAlreadyRegistered)
	})

	t.Run("Auto Generate ID", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		user := domain.User{Login: "autoid", Password: "pass"}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByLogin(ctx, "autoid")
		require.NoError(t, err)
		assert.NotEmpty(t, retrieved.ID)
	})

	t.Run("Preserve Custom ID", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		customID := "custom-id-123"
		user := domain.User{
			ID:       customID,
			Login:    "customid",
			Password: "pass",
		}

		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByLogin(ctx, "customid")
		require.NoError(t, err)
		assert.Equal(t, customID, retrieved.ID)
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		_, err := repo.GetUserByLogin(ctx, "nonexistent")
		assert.ErrorIs(t, err, appErrors.ErrUserNotFound)

		_, err = repo.GetUserByID(ctx, "nonexistent-id")
		assert.ErrorIs(t, err, appErrors.ErrUserNotFound)
	})

	t.Run("Multiple Users", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		users := []string{"user1", "user2", "user3"}
		for _, login := range users {
			err := repo.CreateUser(ctx, domain.User{Login: login, Password: "pass"})
			require.NoError(t, err)
		}

		for _, login := range users {
			user, err := repo.GetUserByLogin(ctx, login)
			assert.NoError(t, err)
			assert.Equal(t, login, user.Login)
		}
	})

	t.Run("User Data Isolation", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		user1 := domain.User{
			ID:       "same-id",
			Login:    "user_a",
			Password: "pass_a",
		}
		user2 := domain.User{
			ID:       "same-id",
			Login:    "user_b",
			Password: "pass_b",
		}

		err := repo.CreateUser(ctx, user1)
		require.NoError(t, err)
		err = repo.CreateUser(ctx, user2)
		require.NoError(t, err)

		retrieved1, err := repo.GetUserByLogin(ctx, "user_a")
		require.NoError(t, err)
		assert.Equal(t, "user_a", retrieved1.Login)

		retrieved2, err := repo.GetUserByLogin(ctx, "user_b")
		require.NoError(t, err)
		assert.Equal(t, "user_b", retrieved2.Login)
	})

	t.Run("Return Copy of User", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		user := domain.User{
			Login:    "original",
			Password: "original_pass",
			Token:    "original_token",
		}

		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		retrieved1, err := repo.GetUserByLogin(ctx, "original")
		require.NoError(t, err)

		retrieved1.Password = "modified_pass"
		retrieved1.Token = "modified_token"

		retrieved2, err := repo.GetUserByLogin(ctx, "original")
		require.NoError(t, err)
		assert.Equal(t, "original_pass", retrieved2.Password)
		assert.Equal(t, "original_token", retrieved2.Token)
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func(index int) {
				user := domain.User{
					Login:    string(rune('a' + index)),
					Password: "pass",
				}
				err := repo.CreateUser(ctx, user)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		for i := 0; i < 5; i++ {
			<-done
		}

		for i := 0; i < 5; i++ {
			login := string(rune('a' + i))
			_, err := repo.GetUserByLogin(ctx, login)
			assert.NoError(t, err)
		}
	})

	t.Run("Empty Repository", func(t *testing.T) {
		repo := NewMemoryAuthRepository()

		_, err := repo.GetUserByLogin(ctx, "anyuser")
		assert.ErrorIs(t, err, appErrors.ErrUserNotFound)

		_, err = repo.GetUserByID(ctx, "anyid")
		assert.ErrorIs(t, err, appErrors.ErrUserNotFound)
	})
}
