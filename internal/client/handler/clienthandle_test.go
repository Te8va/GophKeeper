package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestReadUserChoice(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantChoice  int
		wantError   bool
		description string
	}{
		{
			name:        "Valid choice when not authenticated",
			input:       "1\n",
			wantChoice:  1,
			wantError:   false,
			description: "Should read valid choice when not authenticated",
		},
		{
			name:        "Valid choice when authenticated",
			input:       "3\n",
			wantChoice:  3,
			wantError:   false,
			description: "Should read valid choice when authenticated",
		},
		{
			name:        "Exit choice",
			input:       "0\n",
			wantChoice:  0,
			wantError:   false,
			description: "Should read exit choice",
		},
		{
			name:        "Invalid non-numeric input",
			input:       "abc\n",
			wantChoice:  -1,
			wantError:   true,
			description: "Should return error for non-numeric input",
		},
		{
			name:        "Empty input",
			input:       "\n",
			wantChoice:  -1,
			wantError:   true,
			description: "Should return error for empty input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			go func() {
				defer w.Close()
				fmt.Fprint(w, tt.input)
			}()

			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			choice, err := ReadUserChoice()

			wOut.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, rOut)

			if (err != nil) != tt.wantError {
				t.Errorf("ReadUserChoice() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if choice != tt.wantChoice {
				t.Errorf("ReadUserChoice() = %v, want %v", choice, tt.wantChoice)
			}
		})
	}
}

func TestReadUserChoiceWithScanner(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantChoice  int
		wantError   bool
		description string
	}{
		{
			name:        "Valid numeric input",
			input:       "5",
			wantChoice:  5,
			wantError:   false,
			description: "Should read valid numeric input",
		},
		{
			name:        "Invalid non-numeric input",
			input:       "invalid",
			wantChoice:  -1,
			wantError:   true,
			description: "Should return error for non-numeric input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))

			choice, err := readUserChoiceWithScanner(scanner)

			if (err != nil) != tt.wantError {
				t.Errorf("readUserChoiceWithScanner() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && choice != tt.wantChoice {
				t.Errorf("readUserChoiceWithScanner() = %v, want %v", choice, tt.wantChoice)
			}
		})
	}
}

func readUserChoiceWithScanner(scanner *bufio.Scanner) (int, error) {
	if !scanner.Scan() {
		return -1, fmt.Errorf("ошибка при считывании ввода: %v", scanner.Err())
	}

	action, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return -1, fmt.Errorf("ошибка при преобразовании команды в число: %v", err)
	}

	return action, nil
}

func TestCheckAuth(t *testing.T) {
	tests := []struct {
		name        string
		tokenValue  string
		wantAuth    bool
		wantOutput  string
		description string
	}{
		{
			name:        "Authenticated user",
			tokenValue:  "valid-token",
			wantAuth:    true,
			wantOutput:  "",
			description: "Should return true when token is set",
		},
		{
			name:        "Not authenticated user",
			tokenValue:  "",
			wantAuth:    false,
			wantOutput:  "Ошибка: необходимо сначала войти в систему",
			description: "Should return false and show error when token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldToken := token
			token = tt.tokenValue
			defer func() { token = oldToken }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			gotAuth := checkAuth()

			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if gotAuth != tt.wantAuth {
				t.Errorf("checkAuth() = %v, want %v", gotAuth, tt.wantAuth)
			}
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	tests := []struct {
		name         string
		initialToken string
		wantToken    string
		wantOutput   string
		description  string
	}{
		{
			name:         "Logout with token set",
			initialToken: "some-token",
			wantToken:    "",
			wantOutput:   "✅ Вы вышли из аккаунта",
			description:  "Should clear token and show success message",
		},
		{
			name:         "Logout when already logged out",
			initialToken: "",
			wantToken:    "",
			wantOutput:   "✅ Вы вышли из аккаунта",
			description:  "Should show success message even when no token was set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldToken := token
			token = tt.initialToken
			defer func() { token = oldToken }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logout()

			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if token != tt.wantToken {
				t.Errorf("logout() token = %v, want %v", token, tt.wantToken)
			}
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestChoiceAction(t *testing.T) {
	oldToken := token
	defer func() { token = oldToken }()

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	tests := []struct {
		name       string
		action     int
		tokenValue string
		wantOutput []string
		notOutput  []string
		wantExit   bool
	}{
		{
			name:       "Exit action",
			action:     0,
			tokenValue: "",
			wantOutput: []string{"Выход из программы."},
			wantExit:   true,
		},
		{
			name:       "Register action",
			action:     1,
			tokenValue: "",
			wantOutput: []string{},
			wantExit:   false,
		},
		{
			name:       "Login action",
			action:     2,
			tokenValue: "",
			wantOutput: []string{},
			wantExit:   false,
		},
		{
			name:       "Create data action without auth",
			action:     3,
			tokenValue: "",
			wantOutput: []string{"Ошибка: необходимо сначала войти в систему"},
			wantExit:   false,
		},
		{
			name:       "Get all data action without auth",
			action:     4,
			tokenValue: "",
			wantOutput: []string{"Ошибка: необходимо сначала войти в систему"},
			wantExit:   false,
		},
		{
			name:       "Get data by ID without auth",
			action:     5,
			tokenValue: "",
			wantOutput: []string{"Ошибка: необходимо сначала войти в систему"},
			wantExit:   false,
		},
		{
			name:       "Update data without auth",
			action:     6,
			tokenValue: "",
			wantOutput: []string{"Ошибка: необходимо сначала войти в систему"},
			wantExit:   false,
		},
		{
			name:       "Delete data without auth",
			action:     7,
			tokenValue: "",
			wantOutput: []string{"Ошибка: необходимо сначала войти в систему"},
			wantExit:   false,
		},
		{
			name:       "Logout action",
			action:     8,
			tokenValue: "valid-token",
			wantOutput: []string{"✅ Вы вышли из аккаунта"},
			wantExit:   false,
		},
		{
			name:       "Invalid action",
			action:     999,
			tokenValue: "",
			wantOutput: []string{"Неверная команда: 999"},
			wantExit:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token = tt.tokenValue

			exitCalled := false
			exitCode := 0
			osExit = func(code int) {
				exitCalled = true
				exitCode = code
			}

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			ChoiceAction(tt.action)

			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Test %s: Output should contain '%s', got: %s", tt.name, want, output)
				}
			}

			for _, notWant := range tt.notOutput {
				if strings.Contains(output, notWant) {
					t.Errorf("Test %s: Output should not contain '%s', got: %s", tt.name, notWant, output)
				}
			}

			if tt.name == "Logout action" {
				if token != "" {
					t.Errorf("Test %s: Token should be cleared after logout, got: %s", tt.name, token)
				}
			}

			if tt.wantExit {
				if !exitCalled {
					t.Errorf("Test %s: osExit should have been called", tt.name)
				}
				if exitCode != 0 {
					t.Errorf("Test %s: exit code should be 0, got %d", tt.name, exitCode)
				}
			} else {
				if exitCalled {
					t.Errorf("Test %s: osExit should not have been called", tt.name)
				}
			}
		})
	}
}
