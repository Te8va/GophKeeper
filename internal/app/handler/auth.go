package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Te8va/GophKeeper/internal/app/domain"
	appErrors "github.com/Te8va/GophKeeper/internal/app/errors"
	"github.com/Te8va/GophKeeper/pkg/logger"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_auth.go -package=mocks

type Authorization interface {
	Register(ctx context.Context, login, password string) (string, error)
	Authenticate(ctx context.Context, login, password string) (string, error)
}

type AuthorizationHandler struct {
	aut Authorization
}

func NewAuthorizationHandler(aut Authorization) *AuthorizationHandler {
	return &AuthorizationHandler{aut: aut}
}

func (h *AuthorizationHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var authData domain.AuthorizationData

	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if authData.Login == "" || authData.Password == "" {
		http.Error(w, "Login and password required", http.StatusBadRequest)
		return
	}

	token, err := h.aut.Register(r.Context(), authData.Login, authData.Password)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrAlreadyRegistered):
			http.Error(w, "Login already taken", http.StatusConflict)
		default:
			logger.Logger().Error("Register error", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.setAuthToken(w, token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthorizationHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var authData domain.AuthorizationData

	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.aut.Authenticate(r.Context(), authData.Login, authData.Password)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrWrongPassword),
			errors.Is(err, appErrors.ErrUserNotFound):
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			logger.Logger().Error("Login error", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.setAuthToken(w, token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthorizationHandler) setAuthToken(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})

	w.Header().Set("Authorization", "Bearer "+token)
}
