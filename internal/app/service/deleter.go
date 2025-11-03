package service

import "context"

//go:generate mockgen -source=deleter.go -destination=mocks/delete_mock.gen.go -package=mocks
type DataDeleteServ interface {
	DeleteData(ctx context.Context, dataID, userID string) error
}

func (s *DataService) DeleteData(ctx context.Context, dataID, userID string) error {
	return s.deleter.DeleteData(ctx, dataID, userID)
}
