package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

type FileRepository struct {
	filePath string
	mu       sync.RWMutex
	data     map[string]*domain.DataRecord
}

func NewFileRepository(filePath string) (*FileRepository, error) {
	repo := &FileRepository{
		filePath: filePath,
		data:     make(map[string]*domain.DataRecord),
	}

	if err := repo.loadFromFile(); err != nil {
		return nil, fmt.Errorf("failed to load data from file: %w", err)
	}

	return repo, nil
}

func (r *FileRepository) CreateData(ctx context.Context, userID string, record *domain.DataRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record.ID = uuid.New().String()
	record.UserID = userID

	r.data[record.ID] = record

	return r.saveToFile()
}

func (r *FileRepository) GetDataByID(ctx context.Context, userID, dataID string) (*domain.DataRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.data[dataID]
	if !exists || record.UserID != userID {
		return nil, appErrors.ErrNotFoundData
	}
	return record, nil
}

func (r *FileRepository) GetAllUserData(ctx context.Context, userID string) ([]*domain.DataRecord, error) {
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

func (r *FileRepository) UpdateData(ctx context.Context, userID string, record *domain.DataRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.data[record.ID]
	if !exists || existing.UserID != userID {
		return appErrors.ErrNotFoundData
	}

	record.UserID = userID
	r.data[record.ID] = record

	return r.saveToFile()
}

func (r *FileRepository) DeleteData(ctx context.Context, userID, dataID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.data[dataID]
	if !exists || record.UserID != userID {
		return appErrors.ErrNotFoundData
	}

	delete(r.data, dataID)
	return r.saveToFile()
}

func (r *FileRepository) loadFromFile() error {
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var fileData map[string]*domain.DataRecord
	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	r.data = fileData
	return nil
}

func (r *FileRepository) saveToFile() error {
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
