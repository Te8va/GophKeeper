package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

type FileAuthRepository struct {
	filePath string
	mu       sync.RWMutex
	users    map[string]*domain.User
}

func NewFileAuthRepository(filePath string) (*FileAuthRepository, error) {
	repo := &FileAuthRepository{
		filePath: filePath,
		users:    make(map[string]*domain.User),
	}

	if err := repo.loadFromFile(); err != nil {
		return nil, fmt.Errorf("failed to load auth data from file: %w", err)
	}

	return repo, nil
}

func (r *FileAuthRepository) CreateUser(ctx context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Login]; exists {
		return appErrors.ErrAlreadyRegistered
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	r.users[user.Login] = &user

	return r.saveToFile()
}

func (r *FileAuthRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[login]
	if !exists {
		return nil, appErrors.ErrUserNotFound
	}

	return &domain.User{
		ID:       user.ID,
		Login:    user.Login,
		Password: user.Password,
		Token:    user.Token,
	}, nil
}

func (r *FileAuthRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == userID {
			return &domain.User{
				ID:       user.ID,
				Login:    user.Login,
				Password: user.Password,
				Token:    user.Token,
			}, nil
		}
	}

	return nil, appErrors.ErrUserNotFound
}

func (r *FileAuthRepository) loadFromFile() error {
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read auth file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var fileData map[string]*domain.User
	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal auth data: %w", err)
	}

	r.users = fileData
	return nil
}

func (r *FileAuthRepository) saveToFile() error {
	data, err := json.MarshalIndent(r.users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}
