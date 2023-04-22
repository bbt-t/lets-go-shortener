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
	GetStatistic() (entity.Statistic, error)
}

// NewStorage creates new storage based on config.
func NewStorage(cfg config.Config) (Repository, error) {
	if cfg.StoragePath != "" {
		return newFileStorage(cfg)
	}
	if cfg.BasePath != "" {
		return newDBStorage(cfg)
	}
	return NewMapStorage(cfg)
}
