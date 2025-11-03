package router_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Te8va/GophKeeper/internal/app/config"
	"github.com/Te8va/GophKeeper/internal/app/router"
	"github.com/Te8va/GophKeeper/internal/app/service/mocks"
)

func TestRouter_ConfigurationScenarios(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthorizationServ(ctrl)
	mockSaver := mocks.NewMockDataSaverServ(ctrl)
	mockGetter := mocks.NewMockDataGetterServ(ctrl)
	mockDeleter := mocks.NewMockDataDeleteServ(ctrl)

	testCases := []struct {
		name        string
		cfg         *config.Config
		description string
	}{
		{
			name: "Full configuration with all fields",
			cfg: &config.Config{
				JWTKey:          "jwt-secret-key-123",
				ServerAddress:   ":8080",
				DatabaseDSN:     "postgres://user:pass@localhost:5432/db",
				FileStoragePath: "/tmp/storage.json",
				BaseURL:         "http://localhost:8080",
			},
			description: "Should handle complete configuration",
		},
		{
			name: "Configuration with only required JWT key",
			cfg: &config.Config{
				JWTKey: "minimal-jwt-key",
			},
			description: "Should work with minimal configuration",
		},
		{
			name: "Configuration with very long JWT key",
			cfg: &config.Config{
				JWTKey: "very-long-jwt-key-that-exceeds-typical-length-for-testing-purposes-and-should-still-work-correctly",
			},
			description: "Should handle long JWT keys",
		},
		{
			name: "Configuration with special characters in JWT key",
			cfg: &config.Config{
				JWTKey: "jwt-key-with-special!@#$%^&*()chars",
			},
			description: "Should handle special characters in JWT key",
		},
		{
			name: "Configuration with empty server address",
			cfg: &config.Config{
				JWTKey:        "test-key",
				ServerAddress: "",
			},
			description: "Should handle empty server address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Test case: %s", tc.description)

			r := router.NewRouter(tc.cfg, mockAuth, mockSaver, mockGetter, mockDeleter)

			assert.NotNil(t, r, "Router should be created for configuration: %s", tc.description)
		})
	}
}
