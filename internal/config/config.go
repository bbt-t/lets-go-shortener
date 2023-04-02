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

// Config Application config.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	BaseURL         string `env:"BASE_URL" json:"base_url,omitempty"`
	StoragePath     string `env:"FILE_STORAGE_PATH" json:"storage_path,omitempty"`
	BasePath        string `env:"DATABASE_DSN" json:"base_path,omitempty"`
	DBMigrationPath string
	EnableHTTPS     bool `env:"ENABLE_HTTPS" json:"enable_https,omitempty"`
}

// GetDefaultConfig gets default config.
func GetDefaultConfig() Config {
	return Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://127.0.0.1:8080",
		StoragePath:     "",
		BasePath:        "",
		DBMigrationPath: "file://migrations",
		EnableHTTPS:     false,
	}
}

// Singleton config creation variables.
var (
	cfg  = GetDefaultConfig()
	once sync.Once
)

// GetConfig gets new config from flags or env.
func GetConfig() Config {
	once.Do(parseConfigs)
	return cfg
}

// GetBenchConfig gets config for benchmarks.
func GetBenchConfig() Config {
	return Config{
		DBMigrationPath: "file://../../migrations",
	}
}

// parseConfigs parse values.
func parseConfigs() {
	var (
		fileCfg, envCfg, flagCfg Config
		cfgFilePath              string
	)
	// flag-config
	flag.StringVar(&flagCfg.ServerAddress, "a", cfg.ServerAddress, "Server address")
	flag.StringVar(&flagCfg.BaseURL, "b", cfg.BaseURL, "Base OriginalURL")
	flag.StringVar(&flagCfg.StoragePath, "f", cfg.StoragePath, "Storage path")
	flag.StringVar(&flagCfg.BasePath, "d", cfg.BasePath, "DataBase path")
	flag.BoolVar(&flagCfg.EnableHTTPS, "s", false, "Enable HTTPS")

	// file-config
	flag.StringVar(&cfgFilePath, "config", "", "Config file path")
	flag.StringVar(&cfgFilePath, "c", "", "Config file path")

	flag.Parse()

	if cfgFilePath != "" {
		file, err := os.ReadFile(cfgFilePath)
		if err != nil {
			log.Fatalln("Failed parse config file:", err)
		}

		err = json.Unmarshal(file, &fileCfg)
		if err != nil {
			log.Fatalln("Failed unmarshal config file:", err)
		}
	}

	// env-config
	if err := env.Parse(&envCfg); err != nil {
		log.Fatalln("Failed parse config: ", err)
	}

	// change config by priority.
	cfg.ChangeByPriority(fileCfg)
	cfg.ChangeByPriority(envCfg)
	cfg.ChangeByPriority(flagCfg)
}

// ChangeByPriority changes config by priority.
func (cfg *Config) ChangeByPriority(choiceCfg Config) {
	values := reflect.ValueOf(choiceCfg)
	oldValues := reflect.ValueOf(&cfg).Elem()

	for j := 0; j < values.NumField(); j++ {
		if !values.Field(j).IsZero() {
			oldValues.Field(j).Set(values.Field(j))
		}
	}
}
