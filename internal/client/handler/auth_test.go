package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Te8va/GophKeeper/internal/client/domain"
)

func TestReadCredentials(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLogin   string
		wantPass    string
		description string
	}{
		{
			name:        "Read valid credentials",
			input:       "testuser\ntestpass\n",
			wantLogin:   "testuser",
			wantPass:    "testpass",
			description: "Should read login and password correctly",
		},
		{
			name:        "Read empty credentials",
			input:       "\n\n",
			wantLogin:   "",
			wantPass:    "",
			description: "Should handle empty input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test_input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			tmpfile, err = os.Open(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer tmpfile.Close()

			oldStdin := os.Stdin
			os.Stdin = tmpfile
			defer func() { os.Stdin = oldStdin }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			gotLogin, gotPass := readCredentials()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)

			if gotLogin != tt.wantLogin {
				t.Errorf("readCredentials() login = %v, want %v", gotLogin, tt.wantLogin)
			}
			if gotPass != tt.wantPass {
				t.Errorf("readCredentials() password = %v, want %v", gotPass, tt.wantPass)
			}

			output := buf.String()
			if !strings.Contains(output, "Введите логин:") || !strings.Contains(output, "Введите пароль:") {
				t.Errorf("Expected prompts not found in output: %s", output)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		serverHandler http.HandlerFunc
		wantToken     string
		wantOutput    string
		description   string
	}{
		{
			name:  "Successful registration",
			input: "user1\npass1\n",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "token123"})
				w.WriteHeader(http.StatusOK)
			},
			wantToken:   "token123",
			wantOutput:  "✅ Регистрация успешна!",
			description: "Should register successfully and set token",
		},
		{
			name:  "Registration failure",
			input: "user2\npass2\n",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, "User exists")
			},
			wantToken:   "",
			wantOutput:  "Ошибка регистрации",
			description: "Should show error when registration fails",
		},
		{
			name:  "Registration without cookie",
			input: "user3\npass3\n",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantToken:   "",
			wantOutput:  "✅ Регистрация успешна!",
			description: "Should succeed even without auth cookie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			oldBaseURL := domain.BaseURL
			domain.BaseURL = server.URL
			defer func() { domain.BaseURL = oldBaseURL }()

			tmpfile, err := os.CreateTemp("", "test_input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			tmpfile, err = os.Open(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer tmpfile.Close()

			oldStdin := os.Stdin
			os.Stdin = tmpfile
			defer func() { os.Stdin = oldStdin }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			Register()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}

			token = ""
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		serverHandler http.HandlerFunc
		wantToken     string
		wantOutput    string
		description   string
	}{
		{
			name:  "Successful login",
			input: "user1\npass1\n",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "login-token"})
				w.WriteHeader(http.StatusOK)
			},
			wantToken:   "login-token",
			wantOutput:  "✅ Успешный вход в систему!",
			description: "Should login successfully and set token",
		},
		{
			name:  "Login failure",
			input: "user2\npass2\n",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Invalid credentials")
			},
			wantToken:   "",
			wantOutput:  "Ошибка логина",
			description: "Should show error when login fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			oldBaseURL := domain.BaseURL
			domain.BaseURL = server.URL
			defer func() { domain.BaseURL = oldBaseURL }()

			tmpfile, err := os.CreateTemp("", "test_input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			tmpfile, err = os.Open(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer tmpfile.Close()

			oldStdin := os.Stdin
			os.Stdin = tmpfile
			defer func() { os.Stdin = oldStdin }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			Login()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}

			token = ""
		})
	}
}

func TestEmptyCredentials(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		function    func()
		description string
	}{
		{
			name:        "TC008: Empty credentials for registration",
			input:       "\n\n",
			function:    Register,
			description: "Should exit early without HTTP call",
		},
		{
			name:        "TC009: Empty credentials for login",
			input:       "\n\n",
			function:    Login,
			description: "Should exit early without HTTP call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test_input")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			tmpfile, err = os.Open(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer tmpfile.Close()

			oldStdin := os.Stdin
			os.Stdin = tmpfile
			defer func() { os.Stdin = oldStdin }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			tt.function()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if strings.Contains(output, "Ошибка:") {
				t.Errorf("Should not make HTTP call with empty credentials, output: %s", output)
			}
		})
	}
}
