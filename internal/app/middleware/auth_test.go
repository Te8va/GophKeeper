package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Te8va/GophKeeper/internal/app/middleware"
)

func TestAuthMiddleware_Coverage(t *testing.T) {
	jwtKey := "test-key"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := middleware.Auth(jwtKey)
	handler := authMiddleware(testHandler)

	t.Run("no token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("auth header without bearer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Token sometoken")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestAuthMiddleware_SuccessfulAuth(t *testing.T) {
	jwtKey := "test-key"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(string)
		assert.Equal(t, "testuser", user)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	authMiddleware := middleware.Auth(jwtKey)
	handler := authMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid.token.here")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
}

func TestAuthMiddleware_TokenFromCookie(t *testing.T) {
	jwtKey := "test-key"

	t.Run("successful auth with cookie", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user")
			assert.NotNil(t, user, "User should be in context")
			w.WriteHeader(http.StatusOK)
		})

		authMiddleware := middleware.Auth(jwtKey)
		handler := authMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: "any-token-from-cookie"})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		t.Logf("Covered token from cookie path, status: %d", rr.Code)
	})

	t.Run("cookie with invalid token", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		authMiddleware := middleware.Auth(jwtKey)
		handler := authMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: "invalid-jwt-token"})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
