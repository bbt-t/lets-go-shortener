package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBenchConfig(t *testing.T) {
	cfg := GetBenchConfig()
	assert.Equal(t, Config{
		DBMigrationPath: "file://../../migrations",
	}, cfg)
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	assert.Equal(t, Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://127.0.0.1:8080",
		DBMigrationPath: "file://migrations",
	}, cfg)
}

func TestGetConfig(t *testing.T) {
	os.Args = append(
		os.Args,
		"-a",
		":8088",
		"-b",
		"https://127.0.0.1:8088",
		"-f",
		"storage.db",
		"-d",
		"",
		"-s",
		"-d",
		"postgresql://",
	)
	cfg := GetConfig()

	assert.Equal(t, Config{
		ServerAddress:   ":8088",
		BaseURL:         "https://127.0.0.1:8088",
		StoragePath:     "storage.db",
		BasePath:        "postgresql://",
		DBMigrationPath: "file://migrations",
		EnableHTTPS:     true,
	}, cfg)
}

func TestChangeByPriority(t *testing.T) {
	cfg := GetTestConfig()
	newCfg := Config{
		BaseURL: "https://ololo.com",
	}
	cfg.ChangeByPriority(newCfg)

	assert.Equal(
		t,
		Config{
			ServerAddress:   cfg.ServerAddress,
			BaseURL:         "https://ololo.com",
			StoragePath:     cfg.StoragePath,
			BasePath:        cfg.BasePath,
			DBMigrationPath: cfg.DBMigrationPath,
			EnableHTTPS:     cfg.EnableHTTPS,
		},
		cfg,
	)
}
