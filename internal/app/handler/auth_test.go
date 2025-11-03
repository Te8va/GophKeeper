package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
	"github.com/Te8va/GophKeeper/internal/app/handler"
	"github.com/Te8va/GophKeeper/internal/app/handler/mocks"
)

func TestAuthorizationHandler_RegisterHandler(t *testing.T) {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	tests := []struct {
		name       string
		body       any
		mockReturn string
		mockError  error
		wantStatus int
	}{
		{
			name:       "invalid json",
			body:       "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing fields",
			body:       request{Login: "", Password: ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "already registered",
			body:       request{Login: "user", Password: "pass"},
			mockError:  appErrors.ErrAlreadyRegistered,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "internal error",
			body:       request{Login: "user", Password: "pass"},
			mockError:  errors.New("db fail"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "success",
			body:       request{Login: "user", Password: "pass"},
			mockReturn: "mock-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mocks.NewMockAuthorization(ctrl)
			h := handler.NewAuthorizationHandler(mock)

			var buf bytes.Buffer
			switch v := tt.body.(type) {
			case string:
				buf.WriteString(v)
			default:
				_ = json.NewEncoder(&buf).Encode(v)
			}

			if tt.wantStatus != http.StatusBadRequest {
				authReq := tt.body.(request)
				mock.EXPECT().
					Register(gomock.Any(), authReq.Login, authReq.Password).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", &buf)
			w := httptest.NewRecorder()

			h.RegisterHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusOK {
				require.Equal(t, "Bearer "+tt.mockReturn, res.Header.Get("Authorization"))
				cookie := res.Cookies()[0]
				require.NotEmpty(t, cookie)
				require.Equal(t, "auth_token", cookie.Name)
				require.Equal(t, tt.mockReturn, cookie.Value)
			}
		})
	}
}

func TestAuthorizationHandler_LoginHandler(t *testing.T) {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	tests := []struct {
		name       string
		body       any
		mockReturn string
		mockError  error
		wantStatus int
	}{
		{
			name:       "invalid json",
			body:       "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unauthorized (wrong password)",
			body:       request{Login: "user", Password: "wrong"},
			mockError:  appErrors.ErrWrongPassword,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unauthorized (user not found)",
			body:       request{Login: "ghost", Password: "xxx"},
			mockError:  appErrors.ErrUserNotFound,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "internal error",
			body:       request{Login: "user", Password: "pass"},
			mockError:  errors.New("db fail"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "success",
			body:       request{Login: "user", Password: "pass"},
			mockReturn: "token-123",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mocks.NewMockAuthorization(ctrl)
			h := handler.NewAuthorizationHandler(mock)

			var buf bytes.Buffer
			switch v := tt.body.(type) {
			case string:
				buf.WriteString(v)
			default:
				_ = json.NewEncoder(&buf).Encode(v)
			}

			if tt.wantStatus != http.StatusBadRequest {
				authReq := tt.body.(request)
				mock.EXPECT().
					Authenticate(gomock.Any(), authReq.Login, authReq.Password).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", &buf)
			w := httptest.NewRecorder()

			h.LoginHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusOK {
				require.Equal(t, "Bearer "+tt.mockReturn, res.Header.Get("Authorization"))
				cookie := res.Cookies()[0]
				require.NotEmpty(t, cookie)
				require.Equal(t, "auth_token", cookie.Name)
				require.Equal(t, tt.mockReturn, cookie.Value)
			}
		})
	}
}
