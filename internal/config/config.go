// Package config gets config.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v7"
)

var (
	cfg  = GetDefaultConfig()
	once sync.Once
)

// Config Application config.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	BaseURL         string `env:"BASE_URL" json:"base_url,omitempty"`
	StoragePath     string `env:"FILE_STORAGE_PATH" json:"storage_path,omitempty"`
	BasePath        string `env:"DATABASE_DSN" json:"base_path,omitempty"`
	DBMigrationPath string
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https,omitempty"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	GrpcPort        string `env:"GRPC_RUN_PORT" json:"grpc_port"`
}

// ChangeByPriority changes config by priority.
func (c *Config) ChangeByPriority(newCfg Config) {
	values := reflect.ValueOf(newCfg)
	oldValues := reflect.ValueOf(c).Elem()

	for j := 0; j < values.NumField(); j++ {
		if !values.Field(j).IsZero() {
			oldValues.Field(j).Set(values.Field(j))
		}
	}
}

// GetConfig gets new config from flags, env or file.
func GetConfig() Config {
	once.Do(func() {
		var (
			fileCfg, envCfg, flagCfg Config
			cfgFilePath              string
		)

		flag.StringVar(&flagCfg.ServerAddress, "a", "", "Server address")
		flag.StringVar(&flagCfg.BaseURL, "b", "", "Base URL")
		flag.StringVar(&flagCfg.StoragePath, "f", "", "Storage path")
		flag.StringVar(&flagCfg.BasePath, "d", "", "DataBase path")
		flag.BoolVar(&flagCfg.EnableHTTPS, "s", false, "Enable HTTPS")
		flag.StringVar(&flagCfg.TrustedSubnet, "t", "", "Trusted subnet")
		flag.StringVar(&flagCfg.GrpcPort, "gp", "", "gRPC port")

		flag.StringVar(&cfgFilePath, "c", "", "Config file path")
		flag.StringVar(&cfgFilePath, "config", "", "Config file path")

		flag.Parse()

		if cfgFilePath != "" {
			file, err := os.ReadFile(cfgFilePath)
			if err != nil {
				log.Fatalln("Failed parse config file:", err)
			}

			if err = json.Unmarshal(file, &fileCfg); err != nil {
				log.Fatalln("Failed unmarshal config file:", err)
			}
		}
		// env config.
		err := env.Parse(&envCfg)
		if err != nil {
			log.Fatalln("Failed parse config:", err)
		}
		// change config by priority.
		cfg.ChangeByPriority(fileCfg)
		cfg.ChangeByPriority(envCfg)
		cfg.ChangeByPriority(flagCfg)
	})

	return cfg
}

// GetDefaultConfig gets default config.
func GetDefaultConfig() Config {
	return Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://127.0.0.1:8080",
		DBMigrationPath: "file://migrations",
		EnableHTTPS:     false,
		GrpcPort:        ":3200",
	}
}

// GetBenchConfig gets config for benchmarks.
func GetBenchConfig() Config {
	return Config{
		DBMigrationPath: "file://../../migrations",
	}
}

// GetTestConfig gets config for tests.
func GetTestConfig() Config {
	return Config{
		ServerAddress: ":8080",
		BaseURL:       "http://127.0.0.1:8081",
		StoragePath:   "file_storage.db",
		BasePath:      "mockedDB",
		EnableHTTPS:   false,
		GrpcPort:      ":3200",
	}
}
