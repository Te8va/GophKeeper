package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Te8va/GophKeeper/internal/client/domain"
)

func TestChooseDataType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantType    domain.DataType
		wantOutput  string
		description string
	}{
		{
			name:        "Choose login/password type",
			input:       "1\n",
			wantType:    domain.TypeLoginPassword,
			wantOutput:  "",
			description: "Should return login/password type for choice 1",
		},
		{
			name:        "Choose text data type",
			input:       "2\n",
			wantType:    domain.TypeTextData,
			wantOutput:  "",
			description: "Should return text data type for choice 2",
		},
		{
			name:        "Choose binary data type",
			input:       "3\n",
			wantType:    domain.TypeBinaryData,
			wantOutput:  "",
			description: "Should return binary data type for choice 3",
		},
		{
			name:        "Choose bank card type",
			input:       "4\n",
			wantType:    domain.TypeBankCard,
			wantOutput:  "",
			description: "Should return bank card type for choice 4",
		},
		{
			name:        "Invalid choice",
			input:       "5\n",
			wantType:    "",
			wantOutput:  "Неверный выбор типа данных",
			description: "Should show error for invalid choice",
		},
		{
			name:        "Non-numeric input",
			input:       "abc\n",
			wantType:    "",
			wantOutput:  "Ошибка при преобразовании выбора",
			description: "Should show error for non-numeric input",
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

			gotType := chooseDataType()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if gotType != tt.wantType {
				t.Errorf("chooseDataType() = %v, want %v", gotType, tt.wantType)
			}

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestCreateData(t *testing.T) {
	oldBaseURL := domain.BaseURL
	oldToken := token
	oldStdin := os.Stdin

	defer func() {
		domain.BaseURL = oldBaseURL
		token = oldToken
		os.Stdin = oldStdin
	}()

	tests := []struct {
		name          string
		token         string
		setupInput    func() *os.File
		serverHandler http.HandlerFunc
		wantOutput    []string
		notWantOutput []string
		description   string
	}{
		{
			name:  "Successful data creation",
			token: "valid-token",
			setupInput: func() *os.File {
				r, w, _ := os.Pipe()
				go func() {
					defer w.Close()
					fmt.Fprintln(w, "1")
					fmt.Fprintln(w, "Website login")
					fmt.Fprintln(w, "username: test, password: pass123")
				}()
				return r
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
				}

				if r.Header.Get("Authorization") != "Bearer valid-token" {
					t.Errorf("Expected Authorization: Bearer valid-token, got %s", r.Header.Get("Authorization"))
				}

				var req domain.CreateDataRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("Error decoding request body: %v", err)
				}

				if req.Type != domain.TypeLoginPassword {
					t.Errorf("Expected type %s, got %s", domain.TypeLoginPassword, req.Type)
				}
				if req.Metadata != "Website login" {
					t.Errorf("Expected metadata 'Website login', got '%s'", req.Metadata)
				}
				if req.Data != "username: test, password: pass123" {
					t.Errorf("Expected data 'username: test, password: pass123', got '%s'", req.Data)
				}

				w.WriteHeader(http.StatusCreated)
			},
			wantOutput:    []string{"✅ Запись успешно создана!"},
			notWantOutput: []string{"Ошибка"},
			description:   "Should successfully create data record",
		},
		{
			name:  "Invalid data type choice",
			token: "valid-token",
			setupInput: func() *os.File {
				r, w, _ := os.Pipe()
				go func() {
					defer w.Close()
					fmt.Fprintln(w, "5")
				}()
				return r
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Server should not be called with invalid data type")
			},
			wantOutput:    []string{"Неверный выбор типа данных"},
			notWantOutput: []string{"✅ Запись успешно создана!"},
			description:   "Should handle invalid data type choice",
		},
		{
			name:  "Server returns error",
			token: "valid-token",
			setupInput: func() *os.File {
				r, w, _ := os.Pipe()
				go func() {
					defer w.Close()
					fmt.Fprintln(w, "2")
					fmt.Fprintln(w, "Test note")
					fmt.Fprintln(w, "This is a test note")
				}()
				return r
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error": "Invalid data format"}`)
			},
			wantOutput:    []string{"Ошибка создания записи", "400"},
			notWantOutput: []string{"✅ Запись успешно создана!"},
			description:   "Should handle server error response",
		},
		{
			name:  "Network error",
			token: "valid-token",
			setupInput: func() *os.File {
				r, w, _ := os.Pipe()
				go func() {
					defer w.Close()
					fmt.Fprintln(w, "1")
					fmt.Fprintln(w, "Test")
					fmt.Fprintln(w, "data")
				}()
				return r
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {},
			wantOutput:    []string{"Ошибка:"},
			notWantOutput: []string{"✅ Запись успешно создана!"},
			description:   "Should handle network errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.name != "Network error" {
				server = httptest.NewServer(tt.serverHandler)
				defer server.Close()
				domain.BaseURL = server.URL
			} else {
				domain.BaseURL = "http://invalid-server:9999"
			}

			token = tt.token

			inputFile := tt.setupInput()
			defer inputFile.Close()
			oldStdin := os.Stdin
			os.Stdin = inputFile
			defer func() { os.Stdin = oldStdin }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			createData()

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

			for _, notWant := range tt.notWantOutput {
				if strings.Contains(output, notWant) {
					t.Errorf("Test %s: Output should not contain '%s', got: %s", tt.name, notWant, output)
				}
			}
		})
	}
}

func TestGetAllData(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		serverHandler http.HandlerFunc
		wantOutput    string
		description   string
	}{
		{
			name:  "Successful get all data with records",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				type JSONDataRecord struct {
					ID       string          `json:"id"`
					Type     domain.DataType `json:"type"`
					Metadata string          `json:"metadata"`
					Data     string          `json:"data"`
				}

				records := []JSONDataRecord{
					{
						ID:       "1",
						Type:     domain.TypeLoginPassword,
						Metadata: "Website login",
						Data:     "username: user1, password: pass1",
					},
					{
						ID:       "2",
						Type:     domain.TypeTextData,
						Metadata: "Note",
						Data:     "This is a note",
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(records)
			},
			wantOutput:  "✅ Найдено записей: 2",
			description: "Should display all records successfully",
		},
		{
			name:  "Get all data with no records",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var records []domain.DataRecord
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(records)
			},
			wantOutput:  "Записей не найдено",
			description: "Should show message when no records found",
		},
		{
			name:  "Get all data failure",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Invalid token")
			},
			wantOutput:  "Ошибка получения записей",
			description: "Should show error when request fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			oldBaseURL := domain.BaseURL
			domain.BaseURL = server.URL
			defer func() { domain.BaseURL = oldBaseURL }()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			oldToken := token
			token = tt.token
			defer func() { token = oldToken }()

			getAllData()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}
		})
	}
}
func TestGetDataByID(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		token         string
		serverHandler http.HandlerFunc
		wantOutput    string
		notWantOutput string
		description   string
	}{
		{
			name:  "Successful get by ID",
			input: "123\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, "/data/123") {
					t.Errorf("Expected path to contain /data/123, got %s", r.URL.Path)
				}

				type JSONDataRecord struct {
					ID       string          `json:"id"`
					Type     domain.DataType `json:"type"`
					Metadata string          `json:"metadata"`
					Data     string          `json:"data"`
				}

				record := JSONDataRecord{
					ID:       "123",
					Type:     domain.TypeBankCard,
					Metadata: "Credit card",
					Data:     "Card number: 1234-5678-9012-3456",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(record)
			},
			wantOutput:  "✅ Запись найдена:",
			description: "Should retrieve record by ID successfully",
		},
		{
			name:  "Get by ID not found",
			input: "999\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"error": "Record not found"}`)
			},
			wantOutput:  "Ошибка получения записи",
			description: "Should show error when record not found",
		},
		{
			name:  "Server returns invalid JSON",
			input: "456\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"invalid": json`)
			},
			wantOutput:  "Ошибка декодирования ответа",
			description: "Should handle invalid JSON response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.serverHandler != nil {
				server := httptest.NewServer(tt.serverHandler)
				defer server.Close()

				oldBaseURL := domain.BaseURL
				domain.BaseURL = server.URL
				defer func() { domain.BaseURL = oldBaseURL }()
			} else {
				oldBaseURL := domain.BaseURL
				domain.BaseURL = "http://should-not-call-this-server:8080"
				defer func() { domain.BaseURL = oldBaseURL }()
			}

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			_, err = w.Write([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			w.Close()

			oldStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			oldToken := token
			token = tt.token
			defer func() { token = oldToken }()

			getDataByID()

			wOut.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, rOut)
			output := buf.String()

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}

			if tt.notWantOutput != "" && strings.Contains(output, tt.notWantOutput) {
				t.Errorf("Output should not contain '%s', got: %s", tt.notWantOutput, output)
			}
		})
	}
}

func TestUpdateData(t *testing.T) {
	oldBaseURL := domain.BaseURL
	oldToken := token

	defer func() {
		domain.BaseURL = oldBaseURL
		token = oldToken
	}()

	tests := []struct {
		name          string
		input         string
		token         string
		serverHandler http.HandlerFunc
		wantOutput    []string
		notWantOutput []string
		description   string
	}{
		{
			name:  "Successful update",
			input: "123\n1\nUpdated metadata\nUpdated data\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/data/123") {
					t.Errorf("Expected path to contain /data/123, got %s", r.URL.Path)
				}

				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("Error decoding request body: %v", err)
				}

				if req["type"] != "login_password" {
					t.Errorf("Expected type login_password, got %v", req["type"])
				}
				if req["metadata"] != "Updated metadata" {
					t.Errorf("Expected metadata 'Updated metadata', got '%v'", req["metadata"])
				}
				if req["data"] != "Updated data" {
					t.Errorf("Expected data 'Updated data', got '%v'", req["data"])
				}

				w.WriteHeader(http.StatusCreated)
			},
			wantOutput:    []string{"✅ Запись успешно обновлена!"},
			notWantOutput: []string{"Ошибка"},
			description:   "Should update record successfully",
		},
		{
			name:  "Update failure",
			input: "123\n2\nMeta\nData\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error": "Update failed"}`)
			},
			wantOutput:    []string{"Ошибка обновления записи"},
			notWantOutput: []string{"✅ Запись успешно обновлена!"},
			description:   "Should show error when update fails",
		},
		{
			name:          "Empty ID for update",
			input:         "\n",
			token:         "valid-token",
			serverHandler: nil,
			wantOutput:    []string{"Ошибка при считывании ID: пустой ввод"},
			notWantOutput: []string{"Выберите тип данных:"},
			description:   "Should exit early with empty ID",
		},
		{
			name:          "Invalid data type choice",
			input:         "123\n5\n",
			token:         "valid-token",
			serverHandler: nil,
			wantOutput:    []string{"Неверный выбор типа данных"},
			notWantOutput: []string{"✅ Запись успешно обновлена!"},
			description:   "Should exit early with invalid data type",
		},
		{
			name:          "Non-numeric data type",
			input:         "123\nabc\n",
			token:         "valid-token",
			serverHandler: nil,
			wantOutput:    []string{"Ошибка при преобразовании выбора"},
			description:   "Should handle non-numeric input for data type",
		},
		{
			name:  "Network error",
			input: "123\n1\nTest\ndata\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
			},
			wantOutput:    []string{"Ошибка:"},
			notWantOutput: []string{"✅ Запись успешно обновлена!"},
			description:   "Should handle network errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverHandler != nil && tt.name != "Network error" {
				server = httptest.NewServer(tt.serverHandler)
				defer server.Close()
				domain.BaseURL = server.URL
			} else if tt.name == "Network error" {
				domain.BaseURL = "http://invalid-server:9999"
			} else {
				domain.BaseURL = "http://should-not-call-this-server:8080"
			}

			token = tt.token

			scanner := bufio.NewScanner(strings.NewReader(tt.input))

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			UpdateDataWithScanner(scanner)

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

			for _, notWant := range tt.notWantOutput {
				if strings.Contains(output, notWant) {
					t.Errorf("Test %s: Output should not contain '%s', got: %s", tt.name, notWant, output)
				}
			}
		})
	}
}

func TestDeleteData(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		token         string
		serverHandler http.HandlerFunc
		wantOutput    string
		description   string
	}{
		{
			name:  "Successful delete",
			input: "123\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/data/123") {
					t.Errorf("Expected path to contain /data/123, got %s", r.URL.Path)
				}

				w.WriteHeader(http.StatusOK)
			},
			wantOutput:  "✅ Запись успешно удалена!",
			description: "Should delete record successfully",
		},
		{
			name:  "Delete failure",
			input: "123\n",
			token: "valid-token",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "Record not found")
			},
			wantOutput:  "Ошибка удаления записи",
			description: "Should show error when delete fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.serverHandler != nil {
				server := httptest.NewServer(tt.serverHandler)
				defer server.Close()

				oldBaseURL := domain.BaseURL
				domain.BaseURL = server.URL
				defer func() { domain.BaseURL = oldBaseURL }()
			}

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

			oldToken := token
			token = tt.token
			defer func() { token = oldToken }()

			deleteData()

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Output should contain '%s', got: %s", tt.wantOutput, output)
			}
		})
	}
}
