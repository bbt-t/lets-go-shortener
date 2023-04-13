package usecase

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

// ShortenerService init struct
type ShortenerService struct {
	cfg     config.Config
	storage Repository
}

// NewShortenerService gets new service.
func NewShortenerService(cfg config.Config, s Repository) *ShortenerService {
	return &ShortenerService{
		cfg:     cfg,
		storage: s,
	}
}

// CreateShort creates short url from original.
func (s ShortenerService) CreateShort(userID string, urls ...string) ([]string, error) {
	return s.storage.CreateShort(userID, urls...)
}

// GetOriginal gets original url from short.
func (s ShortenerService) GetOriginal(id string) (string, error) {
	return s.storage.GetOriginal(id)
}

// MarkAsDeleted deletes url.
func (s ShortenerService) MarkAsDeleted(userID string, ids ...string) error {
	return s.storage.MarkAsDeleted(userID, ids...)
}

// GetURLArrayByUser gets all urls.
func (s ShortenerService) GetURLArrayByUser(userID string) ([]entity.URLs, error) {
	return s.storage.GetURLArrayByUser(userID)
}

// PingDB for ping DataBase.
func (s ShortenerService) PingDB() error {
	return s.storage.PingDB()
}

// GetConfig for get cfg.
func (s ShortenerService) GetConfig() config.Config {
	return s.storage.GetConfig()
}

// GetStatistic gets total count of users and urls.
func (s ShortenerService) GetStatistic() (entity.Statistic, error) {
	return s.storage.GetStatistic()
}
