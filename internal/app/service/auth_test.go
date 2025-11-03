package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
	"github.com/Te8va/GophKeeper/internal/app/service/mocks"
)

func TestAuthorization_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthorizationServ(ctrl)
	authService := NewAuthorization(mockRepo, "test_jwt_secret")

	ctx := context.Background()
	login := "user1"
	password := "password123"

	testCases := []struct {
		name          string
		mockSetup     func()
		expectedToken string
		expectedError error
	}{
		{
			name: "successful register",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin(ctx, login).
					Return(nil, appErrors.ErrUserNotFound)
				mockRepo.EXPECT().
					CreateUser(ctx, gomock.Any()).
					Return(nil)
			},
			expectedToken: "mockedToken",
			expectedError: nil,
		},
		{
			name: "already registered",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin(ctx, login).
					Return(&domain.User{}, nil)
			},
			expectedToken: "",
			expectedError: appErrors.ErrAlreadyRegistered,
		},
		{
			name: "unexpected GetUserByLogin error",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin(ctx, login).
					Return(nil, errors.New("db failure"))
			},
			expectedToken: "",
			expectedError: errors.New("service.Register"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			token, err := authService.Register(ctx, login, password)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
			}
		})
	}
}

func TestAuthorization_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthorizationServ(ctrl)
	authService := NewAuthorization(mockRepo, "test_jwt_secret")

	ctx := context.Background()
	login := "user1"
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	token := "mockedToken"

	testCases := []struct {
		name          string
		mockSetup     func()
		expectedToken string
		expectedError error
	}{
		{
			name: "successful authentication",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin(ctx, login).
					Return(&domain.User{
						Login:    login,
						Password: string(hash),
						Token:    token,
					}, nil)
			},
			expectedToken: token,
			expectedError: nil,
		},
		{
			name: "user not found",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin(ctx, login).
					Return(nil, appErrors.ErrUserNotFound)
			},
			expectedToken: "",
			expectedError: appErrors.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			gotToken, err := authService.Authenticate(ctx, login, password)

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
				require.Empty(t, gotToken)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedToken, gotToken)
			}
		})
	}
}
