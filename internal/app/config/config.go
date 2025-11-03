package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress       string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL             string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	FileAuthStoragePath string `env:"FILE_AuthSTORAGE_PATH"`
	PostgresUser        string `env:"POSTGRES_USER"         envDefault:"gophkeeper"`
	PostgresPassword    string `env:"POSTGRES_PASSWORD"     envDefault:"gophkeeper"`
	PostgresDB          string `env:"POSTGRES_DB"     envDefault:"gophkeeper"`
	PostgresPort        int    `env:"POSTGRES_PORT"         envDefault:"5432"`
	DatabaseDSN         string `env:"DATABASE_DSN"`
	JWTKey              string `env:"JWT_KEY"               envDefault:"supermegasecret"`
}

type ConfigFile struct {
	ServerAddress       string `json:"server_address"`
	BaseURL             string `json:"base_url"`
	FileStoragePath     string `json:"file_storage_path"`
	FileAuthStoragePath string `json:"file_authstorage_path"`
	DatabaseDSN         string `json:"database_dsn"`
}

func LoadFromFile(path string) (*ConfigFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfgFile ConfigFile
	if err := json.NewDecoder(f).Decode(&cfgFile); err != nil {
		return nil, err
	}
	return &cfgFile, nil
}

func NewConfig() *Config {

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error parsing environment variables:%v", err)
	}

	serverAddrFlag := flag.String("a", "", "Address to run HTTP server")
	baseURLFlag := flag.String("b", "", "Base URL for short links")
	fileStorageFlag := flag.String("f", "", "Path to storage file")
	fileAuthStorageFlag := flag.String("l", "", "Path to auth storage file")
	databaseDSNFlag := flag.String("d", "", "PostgreSQL connection string")
	configPathFlag := flag.String("c", "", "Path to config file (JSON)")
	configPathFlagLong := flag.String("config", "", "Path to config file (JSON)")

	flag.Parse()

	configPath := ""
	if *configPathFlag != "" {
		configPath = *configPathFlag
	} else if *configPathFlagLong != "" {
		configPath = *configPathFlagLong
	} else if envPath := os.Getenv("CONFIG"); envPath != "" {
		configPath = envPath
	}

	if configPath != "" {
		if cfgFile, err := LoadFromFile(configPath); err == nil {
			cfg.ServerAddress = cfgFile.ServerAddress
			cfg.BaseURL = cfgFile.BaseURL
			cfg.FileStoragePath = cfgFile.FileStoragePath
			cfg.DatabaseDSN = cfgFile.DatabaseDSN
		}
	}

	if *serverAddrFlag != "" {
		cfg.ServerAddress = *serverAddrFlag
	}
	if *baseURLFlag != "" {
		cfg.BaseURL = *baseURLFlag
	}
	if *fileStorageFlag != "" {
		cfg.FileStoragePath = *fileStorageFlag
	}

	if *fileAuthStorageFlag != "" {
		cfg.FileStoragePath = *fileAuthStorageFlag
	}

	if dsnEnv, exists := os.LookupEnv("DATABASE_DSN"); exists && dsnEnv != "" {
		cfg.DatabaseDSN = dsnEnv
	} else if *databaseDSNFlag != "" {
		cfg.DatabaseDSN = *databaseDSNFlag
	}

	return &cfg
}
