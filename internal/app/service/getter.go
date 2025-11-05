package service

import (
	"context"

	"github.com/Te8va/GophKeeper/internal/app/domain"
)

//go:generate mockgen -source=getter.go -destination=mocks/getter_mock.gen.go -package=mocks
type DataGetterServ interface {
	GetDataByID(ctx context.Context, dataID, userID string) (*domain.DataRecord, error)
	GetAllUserData(ctx context.Context, userID string) ([]*domain.DataRecord, error)
}

func (s *DataService) GetDataByID(ctx context.Context, dataID, userID string) (*domain.DataRecord, error) {
	return s.getter.GetDataByID(ctx, dataID, userID)
}

func (s *DataService) GetAllUserData(ctx context.Context, userID string) ([]*domain.DataRecord, error) {
	return s.getter.GetAllUserData(ctx, userID)
}
