package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
	"github.com/Te8va/GophKeeper/pkg/logger"
)

type AuthorizationRepository struct {
	pool *pgxpool.Pool
}

func NewAuthorizationRepository(pool *pgxpool.Pool) *AuthorizationRepository {
	return &AuthorizationRepository{pool: pool}
}

func (r *AuthorizationRepository) CreateUser(ctx context.Context, user domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository.CreateUser: failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logger.Logger().Errorln("CreateUser: failed to rollback transaction:", err)
		}
	}()

	_, err = tx.Exec(ctx, queryInsertUser, user.Login, user.Password, user.Token)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return appErrors.ErrAlreadyRegistered
		}
		return fmt.Errorf("repository.CreateUser: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository.CreateUser: failed to commit transaction: %w", err)
	}

	return nil
}

func (r *AuthorizationRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User

	err := r.pool.QueryRow(ctx, queryGetUserByLogin, login).
		Scan(&user.ID, &user.Login, &user.Password, &user.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.GetUserByLogin: %w", err)
	}

	return &user, nil
}

// вдруг понадобится
func (r *AuthorizationRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User

	err := r.pool.QueryRow(ctx, queryGetUserByID, userID).
		Scan(&user.ID, &user.Login, &user.Password, &user.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.GetUserByID: %w", err)
	}

	return &user, nil
}
