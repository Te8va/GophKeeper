package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/internal/app/config"
)

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantConfig  *config.ConfigFile
		wantErr     bool
	}{
		{
			name: "valid config file",
			fileContent: `{
				"server_address": "localhost:9090",
				"base_url": "http://localhost:9090",
				"file_storage_path": "/tmp/storage.json",
				"file_authstorage_path": "/tmp/auth.json",
				"database_dsn": "postgres://user:pass@localhost:5432/db"
			}`,
			wantConfig: &config.ConfigFile{
				ServerAddress:       "localhost:9090",
				BaseURL:             "http://localhost:9090",
				FileStoragePath:     "/tmp/storage.json",
				FileAuthStoragePath: "/tmp/auth.json",
				DatabaseDSN:         "postgres://user:pass@localhost:5432/db",
			},
			wantErr: false,
		},
		{
			name: "partial config file",
			fileContent: `{
				"server_address": "localhost:9090",
				"base_url": "http://localhost:9090"
			}`,
			wantConfig: &config.ConfigFile{
				ServerAddress: "localhost:9090",
				BaseURL:       "http://localhost:9090",
			},
			wantErr: false,
		},
		{
			name:        "invalid JSON",
			fileContent: `{invalid json}`,
			wantConfig:  nil,
			wantErr:     true,
		},
		{
			name:        "empty file",
			fileContent: "",
			wantConfig:  nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "config_test_*.json")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			if tt.fileContent != "" {
				_, err = tmpFile.WriteString(tt.fileContent)
				require.NoError(t, err)
			}
			tmpFile.Close()

			gotConfig, err := config.LoadFromFile(tmpFile.Name())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotConfig)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantConfig, gotConfig)
			}
		})
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	_, err := config.LoadFromFile("/non/existent/file.json")
	assert.Error(t, err)
}

func TestNewConfig_Priority(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		flags          []string
		configFile     *string
		wantServerAddr string
		wantBaseURL    string
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"SERVER_ADDRESS": "",
				"BASE_URL":       "",
			},
			wantServerAddr: "localhost:8080",
			wantBaseURL:    "http://localhost:8080",
		},
		{
			name: "environment variables override defaults",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env-server:9090",
				"BASE_URL":       "http://env-base-url",
			},
			wantServerAddr: "env-server:9090",
			wantBaseURL:    "http://env-base-url",
		},
		{
			name: "flags override environment variables",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env-server:9090",
				"BASE_URL":       "http://env-base-url",
			},
			flags:          []string{"-a", "flag-server:8081", "-b", "http://flag-base-url"},
			wantServerAddr: "flag-server:8081",
			wantBaseURL:    "http://flag-base-url",
		},
		{
			name: "config file overrides environment variables",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env-server:9090",
				"BASE_URL":       "http://env-base-url",
			},
			configFile: func() *string {
				cfg := `{
					"server_address": "file-server:9091",
					"base_url": "http://file-base-url"
				}`
				return &cfg
			}(),
			wantServerAddr: "file-server:9091",
			wantBaseURL:    "http://file-base-url",
		},
		{
			name: "flags override config file",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env-server:9090",
			},
			configFile: func() *string {
				cfg := `{
					"server_address": "file-server:9091",
					"base_url": "http://file-base-url"
				}`
				return &cfg
			}(),
			flags:          []string{"-a", "flag-server:8081"},
			wantServerAddr: "flag-server:8081",
			wantBaseURL:    "http://file-base-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			oldEnv := make(map[string]string)
			for key := range tt.envVars {
				oldEnv[key] = os.Getenv(key)
			}
			defer func() {
				os.Args = oldArgs
				for key, value := range oldEnv {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			}()

			for key, value := range tt.envVars {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			args := []string{"test"}
			if tt.flags != nil {
				args = append(args, tt.flags...)
			}

			var configFilePath string
			if tt.configFile != nil {
				tmpFile, err := os.CreateTemp("", "config_priority_*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(*tt.configFile)
				require.NoError(t, err)
				tmpFile.Close()

				configFilePath = tmpFile.Name()
				args = append(args, "-c", configFilePath)
			}

			os.Args = args

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg := config.NewConfig()

			assert.Equal(t, tt.wantServerAddr, cfg.ServerAddress)
			assert.Equal(t, tt.wantBaseURL, cfg.BaseURL)
		})
	}
}

func TestNewConfig_DatabaseDSN(t *testing.T) {
	tests := []struct {
		name    string
		envDSN  string
		flagDSN string
		wantDSN string
	}{
		{
			name:    "no DSN set",
			wantDSN: "",
		},
		{
			name:    "environment DSN",
			envDSN:  "postgres://env:pass@localhost:5432/db",
			wantDSN: "postgres://env:pass@localhost:5432/db",
		},
		{
			name:    "flag DSN",
			flagDSN: "postgres://flag:pass@localhost:5432/db",
			wantDSN: "postgres://flag:pass@localhost:5432/db",
		},
		{
			name:    "environment overrides flag",
			envDSN:  "postgres://env:pass@localhost:5432/db",
			flagDSN: "postgres://flag:pass@localhost:5432/db",
			wantDSN: "postgres://env:pass@localhost:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			oldEnvDSN := os.Getenv("DATABASE_DSN")
			defer func() {
				os.Args = oldArgs
				if oldEnvDSN == "" {
					os.Unsetenv("DATABASE_DSN")
				} else {
					os.Setenv("DATABASE_DSN", oldEnvDSN)
				}
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			}()

			if tt.envDSN != "" {
				os.Setenv("DATABASE_DSN", tt.envDSN)
			} else {
				os.Unsetenv("DATABASE_DSN")
			}

			args := []string{"test"}
			if tt.flagDSN != "" {
				args = append(args, "-d", tt.flagDSN)
			}

			os.Args = args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg := config.NewConfig()
			assert.Equal(t, tt.wantDSN, cfg.DatabaseDSN)
		})
	}
}

func TestNewConfig_FileStorage(t *testing.T) {
	tests := []struct {
		name        string
		envStorage  string
		flagStorage string
		configFile  *string
		wantStorage string
	}{
		{
			name:        "no storage path",
			wantStorage: "",
		},
		{
			name:        "environment storage path",
			envStorage:  "/env/storage.json",
			wantStorage: "/env/storage.json",
		},
		{
			name:        "flag storage path",
			flagStorage: "/flag/storage.json",
			wantStorage: "/flag/storage.json",
		},
		{
			name: "config file storage path",
			configFile: func() *string {
				cfg := `{"file_storage_path": "/file/storage.json"}`
				return &cfg
			}(),
			wantStorage: "/file/storage.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			oldEnvStorage := os.Getenv("FILE_STORAGE_PATH")
			defer func() {
				os.Args = oldArgs
				if oldEnvStorage == "" {
					os.Unsetenv("FILE_STORAGE_PATH")
				} else {
					os.Setenv("FILE_STORAGE_PATH", oldEnvStorage)
				}
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			}()

			if tt.envStorage != "" {
				os.Setenv("FILE_STORAGE_PATH", tt.envStorage)
			} else {
				os.Unsetenv("FILE_STORAGE_PATH")
			}

			args := []string{"test"}
			if tt.flagStorage != "" {
				args = append(args, "-f", tt.flagStorage)
			}

			if tt.configFile != nil {
				tmpFile, err := os.CreateTemp("", "config_storage_*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(*tt.configFile)
				require.NoError(t, err)
				tmpFile.Close()

				args = append(args, "-c", tmpFile.Name())
			}

			os.Args = args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg := config.NewConfig()
			assert.Equal(t, tt.wantStorage, cfg.FileStoragePath)
		})
	}
}

func TestNewConfig_DefaultValues(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	os.Args = []string{"test"}
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg := config.NewConfig()

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
	assert.Equal(t, "gophkeeper", cfg.PostgresUser)
	assert.Equal(t, "gophkeeper", cfg.PostgresPassword)
	assert.Equal(t, "gophkeeper", cfg.PostgresDB)
	assert.Equal(t, 5432, cfg.PostgresPort)
	assert.Equal(t, "supermegasecret", cfg.JWTKey)
	assert.Equal(t, "", cfg.FileStoragePath)
	assert.Equal(t, "", cfg.DatabaseDSN)
}

func TestNewConfig_ConfigFileFromEnv(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config_env_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	configContent := `{
		"server_address": "env-config-server:9090",
		"base_url": "http://env-config-base-url"
	}`
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	oldArgs := os.Args
	oldConfigEnv := os.Getenv("CONFIG")
	defer func() {
		os.Args = oldArgs
		if oldConfigEnv == "" {
			os.Unsetenv("CONFIG")
		} else {
			os.Setenv("CONFIG", oldConfigEnv)
		}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	os.Setenv("CONFIG", tmpFile.Name())
	os.Args = []string{"test"}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg := config.NewConfig()

	assert.Equal(t, "env-config-server:9090", cfg.ServerAddress)
	assert.Equal(t, "http://env-config-base-url", cfg.BaseURL)
}
