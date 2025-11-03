package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

type MemoryRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.DataRecord
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]*domain.DataRecord),
	}
}

func (r *MemoryRepository) CreateData(ctx context.Context, userID string, record *domain.DataRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record.ID = uuid.New().String()
	record.UserID = userID
	r.data[record.ID] = record
	return nil
}

func (r *MemoryRepository) GetDataByID(ctx context.Context, userID, dataID string) (*domain.DataRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.data[dataID]
	if !exists || record.UserID != userID {
		return nil, appErrors.ErrNotFoundData
	}
	return record, nil
}

func (r *MemoryRepository) GetAllUserData(ctx context.Context, userID string) ([]*domain.DataRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var records []*domain.DataRecord
	for _, record := range r.data {
		if record.UserID == userID {
			records = append(records, record)
		}
	}
	return records, nil
}

func (r *MemoryRepository) UpdateData(ctx context.Context, userID string, record *domain.DataRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.data[record.ID]
	if !exists || existing.UserID != userID {
		return appErrors.ErrNotFoundData
	}

	record.UserID = userID
	r.data[record.ID] = record
	return nil
}

func (r *MemoryRepository) DeleteData(ctx context.Context, userID, dataID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.data[dataID]
	if !exists || record.UserID != userID {
		return appErrors.ErrNotFoundData
	}

	delete(r.data, dataID)
	return nil
}
