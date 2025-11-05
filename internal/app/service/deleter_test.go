package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Te8va/GophKeeper/internal/app/service"
	"github.com/Te8va/GophKeeper/internal/app/service/mocks"
)

func TestDataService_DeleteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeleter := mocks.NewMockDataDeleteServ(ctrl)

	svc := service.NewDataService(nil, nil, mockDeleter)

	tests := []struct {
		name      string
		dataID    string
		userID    string
		mockSetup func()
		wantErr   bool
	}{
		{
			name:   "successful delete login_password data",
			dataID: "record123",
			userID: "user123",
			mockSetup: func() {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), "record123", "user123").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful delete text data",
			dataID: "record456",
			userID: "user123",
			mockSetup: func() {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), "record456", "user123").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful delete bank card data",
			dataID: "record789",
			userID: "user456",
			mockSetup: func() {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), "record789", "user456").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "delete non-existent data",
			dataID: "nonexistent",
			userID: "user123",
			mockSetup: func() {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), "nonexistent", "user123").
					Return(errors.New("data not found"))
			},
			wantErr: true,
		},
		{
			name:   "delete data with database error",
			dataID: "record123",
			userID: "user123",
			mockSetup: func() {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), "record123", "user123").
					Return(errors.New("database connection failed"))
			},
			wantErr: true,
		},
		{
			name:   "delete data with permission error",
			dataID: "record999",
			userID: "user123",
			mockSetup: func() {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), "record999", "user123").
					Return(errors.New("access denied"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := svc.DeleteData(context.Background(), tt.dataID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
