package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
	"github.com/Te8va/GophKeeper/pkg/jwt"
)

//go:generate mockgen -source=authorization.go -destination=mocks/mock_authorization.go -package=mocks

type AuthorizationServ interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}

type Authorization struct {
	srv    AuthorizationServ
	JWTKey string
}

func NewAuthorization(srv AuthorizationServ, jwtKey string) *Authorization {
	return &Authorization{srv: srv, JWTKey: jwtKey}
}

func (s *Authorization) Register(ctx context.Context, login, password string) (string, error) {
	_, err := s.srv.GetUserByLogin(ctx, login)
	if err == nil {
		return "", appErrors.ErrAlreadyRegistered
	}
	if !errors.Is(err, appErrors.ErrUserNotFound) {
		return "", fmt.Errorf("service.Register: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("service.Register: %w", err)
	}

	tokenStr, err := jwt.CreateJWT(login, []byte(s.JWTKey), time.Now().Add(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("service.Register: %w", err)
	}

	user := domain.User{
		Login:    login,
		Password: string(passwordHash),
		Token:    tokenStr,
	}

	if err := s.srv.CreateUser(ctx, user); err != nil {
		return "", fmt.Errorf("service.Register: %w", err)
	}

	return tokenStr, nil
}

func (s *Authorization) Authenticate(ctx context.Context, login, password string) (string, error) {
	user, err := s.srv.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", appErrors.ErrWrongPassword
	}

	return user.Token, nil
}
