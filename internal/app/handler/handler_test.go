package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
	"github.com/Te8va/GophKeeper/internal/app/handler"
	"github.com/Te8va/GophKeeper/internal/app/handler/mocks"
)

func setupTestHandler(t *testing.T) (*gomock.Controller, *mocks.MockDataSaver, *mocks.MockDataGetter, *mocks.MockDataDelete, *handler.SaverHandler, *handler.GetterHandler, *handler.DeleteHandler) {
	ctrl := gomock.NewController(t)

	mockSaver := mocks.NewMockDataSaver(ctrl)
	mockGetter := mocks.NewMockDataGetter(ctrl)
	mockDeleter := mocks.NewMockDataDelete(ctrl)

	saverHandler := handler.NewSaverHandler(mockSaver)
	getterHandler := handler.NewGetterHandler(mockGetter)
	deleteHandler := handler.NewDeleteHandler(mockDeleter)

	return ctrl, mockSaver, mockGetter, mockDeleter, saverHandler, getterHandler, deleteHandler
}

func TestCreateDataHandler(t *testing.T) {
	ctrl, mockSaver, _, _, saverHandler, _, _ := setupTestHandler(t)
	defer ctrl.Finish()

	testCases := []struct {
		name        string
		contentType string
		body        interface{}
		userID      interface{}
		mockErr     error
		wantCode    int
	}{
		{
			name:        "valid login_password data",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "test account",
				Data:     map[string]interface{}{"login": "test", "password": "pass"},
			},
			userID:   "user123",
			mockErr:  nil,
			wantCode: http.StatusCreated,
		},
		{
			name:        "valid text data",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.TypeTextData,
				Metadata: "important note",
				Data:     "secret text data",
			},
			userID:   "user123",
			mockErr:  nil,
			wantCode: http.StatusCreated,
		},
		{
			name:        "valid bank card data",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.TypeBankCard,
				Metadata: "main card",
				Data:     map[string]interface{}{"number": "1234", "expiry": "12/25", "cvv": "123"},
			},
			userID:   "user456",
			mockErr:  nil,
			wantCode: http.StatusCreated,
		},
		{
			name:        "unauthorized",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "test account",
				Data:     map[string]interface{}{"login": "test", "password": "pass"},
			},
			userID:   nil,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:        "invalid content type",
			contentType: "text/plain",
			body:        "invalid body",
			userID:      "user123",
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "invalid JSON",
			contentType: "application/json",
			body:        "{not:json}",
			userID:      "user123",
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "nil data",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "test account",
				Data:     nil,
			},
			userID:   "user123",
			wantCode: http.StatusBadRequest,
		},
		{
			name:        "empty metadata",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "",
				Data:     map[string]interface{}{"login": "test", "password": "pass"},
			},
			userID:   "user123",
			wantCode: http.StatusBadRequest,
		},
		{
			name:        "invalid data type",
			contentType: "application/json",
			body: domain.CreateDataRequest{
				Type:     domain.DataType("invalid"),
				Metadata: "test account",
				Data:     map[string]interface{}{"login": "test", "password": "pass"},
			},
			userID:   "user123",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			switch v := tc.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tc.contentType)

			if tc.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), domain.UserCtxKey, tc.userID))
			}

			if tc.wantCode == http.StatusCreated {
				mockSaver.EXPECT().
					CreateData(gomock.Any(), tc.userID.(string), gomock.Any()).
					Return(tc.mockErr).
					Times(1)
			}

			w := httptest.NewRecorder()
			saverHandler.CreateDataHandler(w, req)

			require.Equal(t, tc.wantCode, w.Code)
		})
	}
}
func TestUpdateData(t *testing.T) {
	ctrl, mockSaver, _, _, saverHandler, _, _ := setupTestHandler(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		contentType string
		dataID      string
		body        interface{}
		userID      interface{}
		mockSetup   func()
		wantCode    int
	}{
		{
			name:        "successful update login_password data",
			contentType: "application/json",
			dataID:      "record123",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(nil)
			},
			wantCode: http.StatusCreated,
		},
		{
			name:        "successful update text data",
			contentType: "application/json",
			dataID:      "record456",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeTextData,
				Metadata: "updated note",
				Data:     "updated secret text",
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(nil)
			},
			wantCode: http.StatusCreated,
		},
		{
			name:        "successful update bank card data",
			contentType: "application/json",
			dataID:      "record789",
			userID:      "user456",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeBankCard,
				Metadata: "updated card",
				Data:     map[string]interface{}{"number": "1234", "expiry": "12/25", "cvv": "123"},
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user456", gomock.Any()).
					Return(nil)
			},
			wantCode: http.StatusCreated,
		},
		{
			name:        "unauthorized - no user id",
			contentType: "application/json",
			dataID:      "record123",
			userID:      nil,
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {},
			wantCode:  http.StatusUnauthorized,
		},
		{
			name:        "missing data id",
			contentType: "application/json",
			dataID:      "",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {},
			wantCode:  http.StatusBadRequest,
		},
		{
			name:        "invalid content type",
			contentType: "text/plain",
			dataID:      "record123",
			userID:      "user123",
			body:        "plain text",
			mockSetup:   func() {},
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "invalid JSON",
			contentType: "application/json",
			dataID:      "record123",
			userID:      "user123",
			body:        "{not:json}",
			mockSetup:   func() {},
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "nil data",
			contentType: "application/json",
			dataID:      "record123",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     nil,
			},
			mockSetup: func() {},
			wantCode:  http.StatusBadRequest,
		},
		{
			name:        "empty metadata",
			contentType: "application/json",
			dataID:      "record123",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {},
			wantCode:  http.StatusBadRequest,
		},
		{
			name:        "invalid data type",
			contentType: "application/json",
			dataID:      "record123",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.DataType("invalid"),
				Metadata: "updated account",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {},
			wantCode:  http.StatusBadRequest,
		},
		{
			name:        "data not found",
			contentType: "application/json",
			dataID:      "nonexistent",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(appErrors.ErrNotFoundData)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:        "internal server error - database error",
			contentType: "application/json",
			dataID:      "record123",
			userID:      "user123",
			body: domain.UpdateDataRequest{
				Type:     domain.TypeLoginPassword,
				Metadata: "updated account",
				Data:     map[string]interface{}{"login": "newuser", "password": "newpass"},
			},
			mockSetup: func() {
				mockSaver.EXPECT().
					UpdateData(gomock.Any(), "user123", gomock.Any()).
					Return(errors.New("database error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/data/"+tt.dataID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)

			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), domain.UserCtxKey, tt.userID))
			}

			w := httptest.NewRecorder()

			saverHandler.UpdateData(w, req, tt.dataID)

			require.Equal(t, tt.wantCode, w.Code,
				"Test case: %s\nResponse: %s",
				tt.name, w.Body.String())
		})
	}
}

func TestGetDataHandler(t *testing.T) {
	ctrl, _, mockGetter, _, _, getterHandler, _ := setupTestHandler(t)
	defer ctrl.Finish()

	testRecord := &domain.DataRecord{
		ID:       "record123",
		Type:     domain.TypeLoginPassword,
		Metadata: "test account",
		Data:     []byte(`{"login":"test","password":"pass"}`),
	}

	tests := []struct {
		name     string
		dataID   string
		userID   interface{}
		mockData *domain.DataRecord
		mockErr  error
		wantCode int
		wantBody *domain.DataRecord
	}{
		{
			name:     "valid get data by id",
			dataID:   "record123",
			userID:   "user123",
			mockData: testRecord,
			mockErr:  nil,
			wantCode: http.StatusOK,
			wantBody: testRecord,
		},
		{
			name:     "unauthorized - no user id",
			dataID:   "record123",
			userID:   nil,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "unauthorized - empty user id",
			dataID:   "record123",
			userID:   "",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "missing data id",
			dataID:   "",
			userID:   "user123",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "data not found",
			dataID:   "nonexistent",
			userID:   "user123",
			mockData: nil,
			mockErr:  appErrors.ErrNotFoundData,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "internal server error",
			dataID:   "record123",
			userID:   "user123",
			mockData: nil,
			mockErr:  errors.New("database error"),
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/data/"+tt.dataID, nil)

			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), domain.UserCtxKey, tt.userID))
			}

			if tt.wantCode == http.StatusOK || tt.wantCode == http.StatusNotFound {
				mockGetter.EXPECT().
					GetDataByID(gomock.Any(), tt.dataID, tt.userID.(string)).
					Return(tt.mockData, tt.mockErr).
					Times(1)
			}

			w := httptest.NewRecorder()

			getterHandler.GetData(w, req, tt.dataID)

			require.Equal(t, tt.wantCode, w.Code,
				"Test case: %s\nResponse: %s",
				tt.name, w.Body.String())

			if tt.wantCode == http.StatusOK {
				var got domain.DataRecord
				err := json.NewDecoder(w.Body).Decode(&got)
				require.NoError(t, err)
				require.Equal(t, tt.wantBody.ID, got.ID)
				require.Equal(t, tt.wantBody.Type, got.Type)
				require.Equal(t, tt.wantBody.Metadata, got.Metadata)
				require.Equal(t, tt.wantBody.Data, got.Data)
			}
		})
	}
}

func TestGetAllDataHandler(t *testing.T) {
	ctrl, _, mockGetter, _, _, getterHandler, _ := setupTestHandler(t)
	defer ctrl.Finish()

	testRecords := []*domain.DataRecord{
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
	}

	testCases := []struct {
		name     string
		userID   interface{}
		mockData []*domain.DataRecord
		mockErr  error
		wantCode int
		wantBody []*domain.DataRecord
	}{
		{
			name:     "valid get all user data",
			userID:   "user123",
			mockData: testRecords,
			mockErr:  nil,
			wantCode: http.StatusOK,
			wantBody: testRecords,
		},
		{
			name:     "user with no data",
			userID:   "user456",
			mockData: []*domain.DataRecord{},
			mockErr:  nil,
			wantCode: http.StatusOK,
			wantBody: []*domain.DataRecord{},
		},
		{
			name:     "unauthorized",
			userID:   nil,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "internal server error",
			userID:   "user123",
			mockData: nil,
			mockErr:  errors.New("database error"),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/data", nil)

			if tc.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), domain.UserCtxKey, tc.userID))
			}

			if tc.userID != nil {
				mockGetter.EXPECT().
					GetAllUserData(gomock.Any(), tc.userID.(string)).
					Return(tc.mockData, tc.mockErr).
					Times(1)
			}

			w := httptest.NewRecorder()
			getterHandler.GetAllDataHandler(w, req)

			require.Equal(t, tc.wantCode, w.Code)

			if tc.wantCode == http.StatusOK {
				var got []*domain.DataRecord
				err := json.NewDecoder(w.Body).Decode(&got)
				require.NoError(t, err)
				require.Equal(t, len(tc.wantBody), len(got))
				if len(got) > 0 {
					require.Equal(t, tc.wantBody[0].ID, got[0].ID)
					require.Equal(t, tc.wantBody[0].Type, got[0].Type)
				}
			}
		})
	}
}

func TestDeleteData(t *testing.T) {
	ctrl, _, _, mockDeleter, _, _, deleteHandler := setupTestHandler(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		dataID   string
		userID   interface{}
		mockErr  error
		wantCode int
	}{
		{
			name:     "successful delete",
			dataID:   "record123",
			userID:   "user123",
			mockErr:  nil,
			wantCode: http.StatusOK,
		},
		{
			name:     "unauthorized - no user id",
			dataID:   "record123",
			userID:   nil,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "unauthorized - empty user id",
			dataID:   "record123",
			userID:   "",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "missing data id",
			dataID:   "",
			userID:   "user123",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "data not found",
			dataID:   "nonexistent",
			userID:   "user123",
			mockErr:  appErrors.ErrNotFoundData,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "internal server error",
			dataID:   "record123",
			userID:   "user123",
			mockErr:  errors.New("database error"),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/data/"+tt.dataID, nil)

			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), domain.UserCtxKey, tt.userID))
			}

			if tt.wantCode == http.StatusOK || tt.wantCode == http.StatusNotFound || tt.wantCode == http.StatusInternalServerError {
				mockDeleter.EXPECT().
					DeleteData(gomock.Any(), tt.dataID, tt.userID.(string)).
					Return(tt.mockErr).
					Times(1)
			}

			w := httptest.NewRecorder()

			deleteHandler.DeleteData(w, req, tt.dataID)

			require.Equal(t, tt.wantCode, w.Code,
				"Test case: %s\nResponse: %s",
				tt.name, w.Body.String())
		})
	}
}
