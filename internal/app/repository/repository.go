package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
)

type DataRepository struct {
	pool *pgxpool.Pool
}

func NewDataRepository(pool *pgxpool.Pool) *DataRepository {
	return &DataRepository{pool: pool}
}

func (r *DataRepository) CreateData(ctx context.Context, userID string, record *domain.DataRecord) error {
	err := r.pool.QueryRow(ctx, saveDataItemQuery,
		userID,
		record.Type,
		record.Data,
		record.Metadata,
	).Scan(&record.ID)

	if err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	return nil
}

func (r *DataRepository) GetDataByID(ctx context.Context, dataID, userID string) (*domain.DataRecord, error) {
	var record domain.DataRecord

	err := r.pool.QueryRow(ctx, getDataByIDQuery, dataID, userID).Scan(
		&record.ID,
		&record.UserID,
		&record.Type,
		&record.Data,
		&record.Metadata,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("data not found")
		}
		return nil, fmt.Errorf("failed to get data: %w", err)
	}

	return &record, nil
}

func (r *DataRepository) GetAllUserData(ctx context.Context, userID string) ([]*domain.DataRecord, error) {
	rows, err := r.pool.Query(ctx, getAllUserDataQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user data: %w", err)
	}
	defer rows.Close()

	var records []*domain.DataRecord
	for rows.Next() {
		var record domain.DataRecord
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Type,
			&record.Data,
			&record.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data: %w", err)
		}
		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return records, nil
}

func (r *DataRepository) UpdateData(ctx context.Context, userID string, record *domain.DataRecord) error {
	result, err := r.pool.Exec(ctx, updateDataQuery,
		record.Type,
		record.Data,
		record.Metadata,
		record.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return appErrors.ErrNotFoundData
	}

	return nil
}

func (r *DataRepository) DeleteData(ctx context.Context, dataID, userID string) error {
	result, err := r.pool.Exec(ctx, deleteDataQuery, dataID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return appErrors.ErrNotFoundData
	}

	return nil
}
