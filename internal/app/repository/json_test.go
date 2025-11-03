package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

func TestFileRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("CRUD Operations", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "test_data.json")
		repo, err := NewFileRepository(tmpFile)
		require.NoError(t, err)

		record := &domain.DataRecord{
			Type:     domain.TypeTextData,
			Metadata: "Test data",
			Data:     []byte("test content"),
		}

		err = repo.CreateData(ctx, "user1", record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.Equal(t, "user1", record.UserID)

		retrieved, err := repo.GetDataByID(ctx, "user1", record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, retrieved.ID)
		assert.Equal(t, "Test data", retrieved.Metadata)

		allData, err := repo.GetAllUserData(ctx, "user1")
		require.NoError(t, err)
		assert.Len(t, allData, 1)
		assert.Equal(t, record.ID, allData[0].ID)

		record.Metadata = "Updated data"
		err = repo.UpdateData(ctx, "user1", record)
		require.NoError(t, err)

		updated, err := repo.GetDataByID(ctx, "user1", record.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated data", updated.Metadata)

		err = repo.DeleteData(ctx, "user1", record.ID)
		require.NoError(t, err)

		_, err = repo.GetDataByID(ctx, "user1", record.ID)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)
	})

	t.Run("User Isolation", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "isolation.json")
		repo, err := NewFileRepository(tmpFile)
		require.NoError(t, err)

		record1 := &domain.DataRecord{Metadata: "User1 data", Data: []byte("user1")}
		record2 := &domain.DataRecord{Metadata: "User2 data", Data: []byte("user2")}

		err = repo.CreateData(ctx, "user1", record1)
		require.NoError(t, err)
		err = repo.CreateData(ctx, "user2", record2)
		require.NoError(t, err)

		user1Data, err := repo.GetAllUserData(ctx, "user1")
		require.NoError(t, err)
		assert.Len(t, user1Data, 1)
		assert.Equal(t, "User1 data", user1Data[0].Metadata)

		user2Data, err := repo.GetAllUserData(ctx, "user2")
		require.NoError(t, err)
		assert.Len(t, user2Data, 1)
		assert.Equal(t, "User2 data", user2Data[0].Metadata)

		_, err = repo.GetDataByID(ctx, "user1", record2.ID)
		assert.ErrorIs(t, err, appErrors.ErrNotFoundData)
	})

	t.Run("Data Persistence", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "persistence.json")

		repo1, err := NewFileRepository(tmpFile)
		require.NoError(t, err)

		record := &domain.DataRecord{Metadata: "Persistent", Data: []byte("data")}
		err = repo1.CreateData(ctx, "user1", record)
		require.NoError(t, err)

		repo2, err := NewFileRepository(tmpFile)
		require.NoError(t, err)

		retrieved, err := repo2.GetDataByID(ctx, "user1", record.ID)
		require.NoError(t, err)
		assert.Equal(t, "Persistent", retrieved.Metadata)
	})

	t.Run("Multiple Records", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "multiple.json")
		repo, err := NewFileRepository(tmpFile)
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			record := &domain.DataRecord{
				Metadata: string(rune('A' + i)),
				Data:     []byte{byte(i)},
			}
			err := repo.CreateData(ctx, "user1", record)
			require.NoError(t, err)
		}

		allData, err := repo.GetAllUserData(ctx, "user1")
		require.NoError(t, err)
		assert.Len(t, allData, 3)
	})

}
