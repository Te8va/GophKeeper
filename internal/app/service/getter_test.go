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

func TestDataService_GetDataByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := mocks.NewMockDataGetterServ(ctrl)

	svc := service.NewDataService(nil, mockGetter, nil)

	tests := []struct {
		name      string
		dataID    string
		userID    string
		mockSetup func()
		wantData  *domain.DataRecord
		wantErr   bool
	}{
		{
			name:   "successful get login_password data by id",
			dataID: "record123",
			userID: "user123",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetDataByID(gomock.Any(), "record123", "user123").
					Return(&domain.DataRecord{
						ID:       "record123",
						Type:     domain.TypeLoginPassword,
						Metadata: "test account",
						Data:     []byte(`{"login":"test","password":"pass"}`),
					}, nil)
			},
			wantData: &domain.DataRecord{
				ID:       "record123",
				Type:     domain.TypeLoginPassword,
				Metadata: "test account",
				Data:     []byte(`{"login":"test","password":"pass"}`),
			},
			wantErr: false,
		},
		{
			name:   "successful get text data by id",
			dataID: "record456",
			userID: "user123",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetDataByID(gomock.Any(), "record456", "user123").
					Return(&domain.DataRecord{
						ID:       "record456",
						Type:     domain.TypeTextData,
						Metadata: "important note",
						Data:     []byte("secret text data"),
					}, nil)
			},
			wantData: &domain.DataRecord{
				ID:       "record456",
				Type:     domain.TypeTextData,
				Metadata: "important note",
				Data:     []byte("secret text data"),
			},
			wantErr: false,
		},
		{
			name:   "get non-existent data by id",
			dataID: "nonexistent",
			userID: "user123",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetDataByID(gomock.Any(), "nonexistent", "user123").
					Return(nil, errors.New("data not found"))
			},
			wantData: nil,
			wantErr:  true,
		},
		{
			name:   "get data with database error",
			dataID: "record123",
			userID: "user123",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetDataByID(gomock.Any(), "record123", "user123").
					Return(nil, errors.New("database connection failed"))
			},
			wantData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			data, err := svc.GetDataByID(context.Background(), tt.dataID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, data)
			}
		})
	}
}

func TestDataService_GetAllUserData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := mocks.NewMockDataGetterServ(ctrl)

	svc := service.NewDataService(nil, mockGetter, nil)

	tests := []struct {
		name      string
		userID    string
		mockSetup func()
		wantData  []*domain.DataRecord
		wantErr   bool
	}{
		{
			name:   "successful get all user data",
			userID: "user123",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetAllUserData(gomock.Any(), "user123").
					Return([]*domain.DataRecord{
						{
							ID:       "record123",
							Type:     domain.TypeLoginPassword,
							Metadata: "test account",
							Data:     []byte(`{"login":"test","password":"pass"}`),
						},
						{
							ID:       "record456",
							Type:     domain.TypeTextData,
							Metadata: "important note",
							Data:     []byte("secret text data"),
						},
					}, nil)
			},
			wantData: []*domain.DataRecord{
				{
					ID:       "record123",
					Type:     domain.TypeLoginPassword,
					Metadata: "test account",
					Data:     []byte(`{"login":"test","password":"pass"}`),
				},
				{
					ID:       "record456",
					Type:     domain.TypeTextData,
					Metadata: "important note",
					Data:     []byte("secret text data"),
				},
			},
			wantErr: false,
		},
		{
			name:   "get all data for user with no records",
			userID: "user456",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetAllUserData(gomock.Any(), "user456").
					Return([]*domain.DataRecord{}, nil)
			},
			wantData: []*domain.DataRecord{},
			wantErr:  false,
		},
		{
			name:   "get all data with database error",
			userID: "user123",
			mockSetup: func() {
				mockGetter.EXPECT().
					GetAllUserData(gomock.Any(), "user123").
					Return(nil, errors.New("database connection failed"))
			},
			wantData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			data, err := svc.GetAllUserData(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, data)
			}
		})
	}
}
