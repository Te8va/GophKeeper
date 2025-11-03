package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	"github.com/Te8va/GophKeeper/internal/app/service"
	"github.com/Te8va/GophKeeper/internal/app/service/mocks"
)

func TestDataService_CreateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := mocks.NewMockDataSaverServ(ctrl)

	svc := service.NewDataService(mockSaver, nil, nil)

	tests := []struct {
		name      string
		userID    string
		record    *domain.DataRecord
		mockSetup func()
		wantErr   bool
	}{
		{
			name:   "successful create login_password data",
			userID: "user123",
			record: &domain.DataRecord{
				Type:     domain.TypeLoginPassword,
				Metadata: "test account",
				Data:     []byte(`{"login":"test","password":"pass"}`),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					CreateData(gomock.Any(), "user123", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful create text data",
			userID: "user123",
			record: &domain.DataRecord{
				Type:     domain.TypeTextData,
				Metadata: "important note",
				Data:     []byte("secret text data"),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					CreateData(gomock.Any(), "user123", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful create bank card data",
			userID: "user456",
			record: &domain.DataRecord{
				Type:     domain.TypeBankCard,
				Metadata: "main card",
				Data:     []byte(`{"number":"1234","expiry":"12/25","cvv":"123"}`),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					CreateData(gomock.Any(), "user456", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "create data with database error",
			userID: "user123",
			record: &domain.DataRecord{
				Type:     domain.TypeLoginPassword,
				Metadata: "test account",
				Data:     []byte(`{"login":"test","password":"pass"}`),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					CreateData(gomock.Any(), "user123", gomock.Any()).
					Return(errors.New("database connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := svc.CreateData(context.Background(), tt.userID, tt.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataService_UpdateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := mocks.NewMockDataSaverServ(ctrl)

	svc := service.NewDataService(mockSaver, nil, nil)

	tests := []struct {
		name      string
		userID    string
		record    *domain.DataRecord
		mockSetup func()
		wantErr   bool
	}{
		{
			name:   "successful update login_password data",
			userID: "user123",
			record: &domain.DataRecord{
				ID:       "record123",
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     []byte(`{"login":"newuser","password":"newpass"}`),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful update text data",
			userID: "user123",
			record: &domain.DataRecord{
				ID:       "record456",
				Type:     domain.TypeTextData,
				Metadata: "updated note",
				Data:     []byte("updated secret text"),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful update binary data",
			userID: "user789",
			record: &domain.DataRecord{
				ID:       "record789",
				Type:     domain.TypeBinaryData,
				Metadata: "updated file",
				Data:     []byte("updated binary content"),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user789", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "update data with database error",
			userID: "user123",
			record: &domain.DataRecord{
				ID:       "record123",
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     []byte(`{"login":"newuser","password":"newpass"}`),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(errors.New("update failed"))
			},
			wantErr: true,
		},
		{
			name:   "update non-existent record",
			userID: "user123",
			record: &domain.DataRecord{
				ID:       "nonexistent",
				Type:     domain.TypeLoginPassword,
				Metadata: "non-existent record",
				Data:     []byte(`{"login":"test","password":"pass"}`),
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(errors.New("record not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := svc.UpdateData(context.Background(), tt.userID, tt.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataService_NewDataService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := mocks.NewMockDataSaverServ(ctrl)
	mockGetter := mocks.NewMockDataGetterServ(ctrl)
	mockDeleter := mocks.NewMockDataDeleteServ(ctrl)

	svc := service.NewDataService(mockSaver, mockGetter, mockDeleter)

	assert.NotNil(t, svc, "DataService should not be nil")
}
