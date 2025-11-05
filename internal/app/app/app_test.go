package app_test

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Te8va/GophKeeper/internal/app/app"
	"github.com/Te8va/GophKeeper/internal/app/config"
	"github.com/Te8va/GophKeeper/internal/app/service/mocks"
)

func TestNewApp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		args       []string
		wantErr    bool
	}{
		{
			name: "successful app creation with memory storage",
			setupEnv: func() {
				os.Unsetenv("DATABASE_DSN")
				os.Unsetenv("FILE_STORAGE_PATH")
			},
			cleanupEnv: func() {},
			args:       []string{"test"},
			wantErr:    false,
		},
		{
			name: "successful app creation with file storage env",
			setupEnv: func() {
				os.Unsetenv("DATABASE_DSN")
				os.Setenv("FILE_STORAGE_PATH", "/tmp/test_storage.json")
			},
			cleanupEnv: func() {
				os.Unsetenv("FILE_STORAGE_PATH")
			},
			args:    []string{"test"},
			wantErr: false,
		},
		{
			name: "successful app creation with file storage flag",
			setupEnv: func() {
				os.Unsetenv("DATABASE_DSN")
				os.Unsetenv("FILE_STORAGE_PATH")
			},
			cleanupEnv: func() {},
			args:       []string{"test", "-f", "/tmp/test_storage.json"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldDatabaseDSN := os.Getenv("DATABASE_DSN")
			oldFileStoragePath := os.Getenv("FILE_STORAGE_PATH")

			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ContinueOnError)

			os.Args = tt.args

			defer func() {
				if oldDatabaseDSN != "" {
					os.Setenv("DATABASE_DSN", oldDatabaseDSN)
				} else {
					os.Unsetenv("DATABASE_DSN")
				}
				if oldFileStoragePath != "" {
					os.Setenv("FILE_STORAGE_PATH", oldFileStoragePath)
				} else {
					os.Unsetenv("FILE_STORAGE_PATH")
				}
			}()

			tt.setupEnv()
			defer tt.cleanupEnv()

			application, err := app.NewApp()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, application)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, application)
				assert.NotNil(t, application.Config())
				assert.NotNil(t, application.Logger())
				assert.NotNil(t, application.Saver())
				assert.NotNil(t, application.Getter())
				assert.NotNil(t, application.Deleter())
				assert.NotNil(t, application.Auth())
				assert.NotNil(t, application.HTTPServer())
			}
		})
	}
}
func TestApp_InitStorage(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.Config
		expectedSaver  bool
		expectedGetter bool
		expectedAuth   bool
		wantErr        bool
	}{
		{
			name: "init memory storage when no storage configured",
			cfg: &config.Config{
				DatabaseDSN:     "",
				FileStoragePath: "",
			},
			expectedSaver:  true,
			expectedGetter: true,
			expectedAuth:   true,
			wantErr:        false,
		},
		{
			name: "init file storage when file path provided",
			cfg: &config.Config{
				DatabaseDSN:     "",
				FileStoragePath: "/tmp/test_storage.json",
			},
			expectedSaver:  true,
			expectedGetter: true,
			expectedAuth:   true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testApp := &app.App{}
			testApp.SetConfig(tt.cfg)

			logger, _ := zap.NewDevelopment()
			sugar := logger.Sugar()
			testApp.SetLogger(sugar)

			err := testApp.InitStorage()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedSaver {
					assert.NotNil(t, testApp.Saver())
				}
				if tt.expectedGetter {
					assert.NotNil(t, testApp.Getter())
				}
				if tt.expectedAuth {
					assert.NotNil(t, testApp.Auth())
				}
				assert.NotNil(t, testApp.Deleter())
			}
		})
	}
}

func TestApp_InitServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := mocks.NewMockDataSaverServ(ctrl)
	mockGetter := mocks.NewMockDataGetterServ(ctrl)
	mockDeleter := mocks.NewMockDataDeleteServ(ctrl)
	mockAuth := mocks.NewMockAuthorizationServ(ctrl)

	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}

	testApp := &app.App{}
	testApp.SetConfig(cfg)
	testApp.SetSaver(mockSaver)
	testApp.SetGetter(mockGetter)
	testApp.SetDeleter(mockDeleter)
	testApp.SetAuth(mockAuth)

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	testApp.SetLogger(sugar)

	testApp.InitServer()

	assert.NotNil(t, testApp.HTTPServer())
	assert.Equal(t, "localhost:8080", testApp.HTTPServer().Addr)
	assert.NotNil(t, testApp.HTTPServer().Handler)
}

func TestApp_StoragePriority(t *testing.T) {
	tests := []struct {
		name            string
		databaseDSN     string
		fileStoragePath string
		wantErr         bool
	}{
		{
			name:            "use file storage when no database",
			databaseDSN:     "",
			fileStoragePath: "/tmp/test_storage.json",
			wantErr:         false,
		},
		{
			name:            "use memory storage when no other options",
			databaseDSN:     "",
			fileStoragePath: "",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			args := []string{"test"}
			if tt.fileStoragePath != "" {
				args = append(args, "-f", tt.fileStoragePath)
			}
			os.Args = args

			application, err := app.NewApp()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, application)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, application)
				assert.NotNil(t, application.Saver())
				assert.NotNil(t, application.Getter())
				assert.NotNil(t, application.Deleter())
				assert.NotNil(t, application.Auth())

				if tt.fileStoragePath != "" {
					os.Remove(tt.fileStoragePath)
					os.Remove("/tmp/auth.json")
				}
			}
		})
	}
}

func TestApp_SettersAndGetters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := mocks.NewMockDataSaverServ(ctrl)
	mockGetter := mocks.NewMockDataGetterServ(ctrl)
	mockDeleter := mocks.NewMockDataDeleteServ(ctrl)
	mockAuth := mocks.NewMockAuthorizationServ(ctrl)
	mockServer := &http.Server{}
	testCfg := &config.Config{ServerAddress: "test:8080"}
	testLogger, _ := zap.NewDevelopment()
	testSugar := testLogger.Sugar()

	testApp := &app.App{}

	testApp.SetConfig(testCfg)
	testApp.SetLogger(testSugar)
	testApp.SetSaver(mockSaver)
	testApp.SetGetter(mockGetter)
	testApp.SetDeleter(mockDeleter)
	testApp.SetAuth(mockAuth)
	testApp.SetHTTPServer(mockServer)

	assert.Equal(t, testCfg, testApp.Config())
	assert.Equal(t, testSugar, testApp.Logger())
	assert.Equal(t, mockSaver, testApp.Saver())
	assert.Equal(t, mockGetter, testApp.Getter())
	assert.Equal(t, mockDeleter, testApp.Deleter())
	assert.Equal(t, mockAuth, testApp.Auth())
	assert.Equal(t, mockServer, testApp.HTTPServer())
}

func TestApp_MemoryStorageInitialization(t *testing.T) {
	testApp := &app.App{}
	testApp.SetConfig(&config.Config{
		DatabaseDSN:     "",
		FileStoragePath: "",
	})

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	testApp.SetLogger(sugar)

	err := testApp.InitStorage()
	assert.NoError(t, err)

	assert.NotNil(t, testApp.Saver())
	assert.NotNil(t, testApp.Getter())
	assert.NotNil(t, testApp.Deleter())
	assert.NotNil(t, testApp.Auth())
}

func TestApp_Run_CoverEdgeCases(t *testing.T) {
	t.Run("waitGroupChan path", func(t *testing.T) {
		testServer := &http.Server{
			Addr:    "localhost:0",
			Handler: http.NewServeMux(),
		}

		cfg := &config.Config{ServerAddress: "localhost:0"}
		logger, _ := zap.NewDevelopment()

		testApp := &app.App{}
		testApp.SetConfig(cfg)
		testApp.SetLogger(logger.Sugar())
		testApp.SetHTTPServer(testServer)

		go testApp.Run()
		time.Sleep(100 * time.Millisecond)

		testServer.Shutdown(context.Background())
	})

	t.Run("timeout path", func(t *testing.T) {
		testHandler := http.NewServeMux()
		testHandler.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second)
		})

		testServer := &http.Server{
			Addr:    "localhost:0",
			Handler: testHandler,
		}

		cfg := &config.Config{ServerAddress: "localhost:0"}
		logger, _ := zap.NewDevelopment()

		testApp := &app.App{}
		testApp.SetConfig(cfg)
		testApp.SetLogger(logger.Sugar())
		testApp.SetHTTPServer(testServer)

		go testApp.Run()
		time.Sleep(100 * time.Millisecond)

		go http.Get("http://" + testServer.Addr + "/slow")
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		testServer.Shutdown(ctx)
	})
}

func TestApp_InitPostgresStorage_BasicCoverage(t *testing.T) {
	tests := []struct {
		name        string
		databaseDSN string
		expectError bool
	}{
		{
			name:        "invalid DSN - covers first error path with different error",
			databaseDSN: "invalid-dsn",
			expectError: true,
		},
		{
			name:        "valid DSN format - will fail on migrations but covers more code",
			databaseDSN: "postgres://user:pass@localhost:5432/db",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg := &config.Config{
				DatabaseDSN: tt.databaseDSN,
			}

			logger, _ := zap.NewDevelopment()
			sugar := logger.Sugar()

			testApp := &app.App{}
			testApp.SetConfig(cfg)
			testApp.SetLogger(sugar)

			err := testApp.InitPostgresStorage(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.databaseDSN == "" {
					assert.Contains(t, err.Error(), "failed to initialize migrations")
				} else {
					assert.NotNil(t, err)
				}
				t.Logf("Test '%s': got expected error: %v", tt.name, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NotNil(t, testApp.Logger())
			assert.NotNil(t, testApp.Config())
		})
	}
}
