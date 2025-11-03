package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

type MemoryAuthRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewMemoryAuthRepository() *MemoryAuthRepository {
	return &MemoryAuthRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *MemoryAuthRepository) CreateUser(ctx context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Login]; exists {
		return appErrors.ErrAlreadyRegistered
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	r.users[user.Login] = &user
	return nil
}

func (r *MemoryAuthRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
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

func (r *MemoryAuthRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
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
