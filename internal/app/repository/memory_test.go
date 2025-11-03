package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

func TestMemoryRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Create and Get Data", func(t *testing.T) {
		repo := NewMemoryRepository()

		record := &domain.DataRecord{
			Type:     domain.TypeTextData,
			Metadata: "Test data",
			Data:     []byte("test content"),
		}

		err := repo.CreateData(ctx, "user1", record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.Equal(t, "user1", record.UserID)

		retrieved, err := repo.GetDataByID(ctx, "user1", record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, retrieved.ID)
		assert.Equal(t, "Test data", retrieved.Metadata)
		assert.Equal(t, []byte("test content"), retrieved.Data)
	})

	t.Run("Get All User Data", func(t *testing.T) {
		repo := NewMemoryRepository()

		record1 := &domain.DataRecord{Metadata: "First", Data: []byte("1")}
		record2 := &domain.DataRecord{Metadata: "Second", Data: []byte("2")}

		err := repo.CreateData(ctx, "user1", record1)
		require.NoError(t, err)
		err = repo.CreateData(ctx, "user1", record2)
		require.NoError(t, err)

		allData, err := repo.GetAllUserData(ctx, "user1")
		require.NoError(t, err)
		assert.Len(t, allData, 2)
	})

	t.Run("Update Data", func(t *testing.T) {
		repo := NewMemoryRepository()

		record := &domain.DataRecord{Metadata: "Original", Data: []byte("data")}
		err := repo.CreateData(ctx, "user1", record)
		require.NoError(t, err)

		record.Metadata = "Updated"
		record.Data = []byte("new data")
		err = repo.UpdateData(ctx, "user1", record)
		require.NoError(t, err)

		updated, err := repo.GetDataByID(ctx, "user1", record.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Metadata)
		assert.Equal(t, []byte("new data"), updated.Data)
	})

	t.Run("Delete Data", func(t *testing.T) {
		repo := NewMemoryRepository()

		record := &domain.DataRecord{Metadata: "To delete", Data: []byte("data")}
		err := repo.CreateData(ctx, "user1", record)
		require.NoError(t, err)

		err = repo.DeleteData(ctx, "user1", record.ID)
		require.NoError(t, err)

		_, err = repo.GetDataByID(ctx, "user1", record.ID)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)
	})

	t.Run("Error Cases - Not Found", func(t *testing.T) {
		repo := NewMemoryRepository()

		_, err := repo.GetDataByID(ctx, "user1", "nonexistent")
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)

		record := &domain.DataRecord{ID: "nonexistent", Metadata: "Test"}
		err = repo.UpdateData(ctx, "user1", record)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)

		err = repo.DeleteData(ctx, "user1", "nonexistent")
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)
	})

	t.Run("Error Cases - Wrong User", func(t *testing.T) {
		repo := NewMemoryRepository()

		record := &domain.DataRecord{Metadata: "User1 data"}
		err := repo.CreateData(ctx, "user1", record)
		require.NoError(t, err)

		_, err = repo.GetDataByID(ctx, "user2", record.ID)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)

		err = repo.UpdateData(ctx, "user2", record)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)

		err = repo.DeleteData(ctx, "user2", record.ID)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)
	})

	t.Run("Empty Repository", func(t *testing.T) {
		repo := NewMemoryRepository()

		data, err := repo.GetAllUserData(ctx, "user1")
		require.NoError(t, err)
		assert.Empty(t, data)

		_, err = repo.GetDataByID(ctx, "user1", "any")
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)
	})
}
