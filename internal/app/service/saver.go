package service

import (
	"context"

	"github.com/Te8va/GophKeeper/internal/app/domain"
)

type DataService struct {
	saver   DataSaverServ
	getter  DataGetterServ
	deleter DataDeleteServ
}

func NewDataService(saver DataSaverServ, getter DataGetterServ, deleter DataDeleteServ) *DataService {
	return &DataService{saver: saver, getter: getter, deleter: deleter}
}

//go:generate mockgen -source=saver.go -destination=mocks/saver_mock.gen.go -package=mocks
type DataSaverServ interface {
	CreateData(ctx context.Context, userID string, req *domain.DataRecord) error
	UpdateData(ctx context.Context, userID string, req *domain.DataRecord) error
}

func (s *DataService) CreateData(ctx context.Context, userID string, req *domain.DataRecord) error {
	return s.saver.CreateData(ctx, userID, req)
}

func (s *DataService) UpdateData(ctx context.Context, userID string, req *domain.DataRecord) error {
	return s.saver.UpdateData(ctx, userID, req)
}
