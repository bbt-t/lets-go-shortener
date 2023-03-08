// Repository implementation package.

package storage

import (
	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
)

// Repository is an interface that describes storage.
type Repository interface {
	CreateShort(userID string, urls ...string) ([]string, error)
	GetOriginal(id string) (string, error)
	MarkAsDeleted(userID string, ids ...string) error
	GetURLArrayByUser(userID string) ([]entity.URLs, error)
	PingDB() error
	GetConfig() config.Config
}

// NewStorage creates new storage based on config.
func NewStorage(cfg config.Config) (Repository, error) {
	if cfg.StoragePath != "" {
		return NewFileStorage(cfg)
	}
	if cfg.BasePath != "" {
		return NewDBStorage(cfg)
	}
	return NewMapStorage(cfg)
}
