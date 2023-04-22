package storage

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
)

// MapStorage is storage that storages in map.
type MapStorage struct {
	Cfg       config.Config
	Locations map[string]string
	Users     map[string][]string
	Deleted   map[string]bool
	*sync.Mutex
}

// NewMapStorage creates new map storage.
func NewMapStorage(cfg config.Config) (*MapStorage, error) {
	return &MapStorage{
		Locations: make(map[string]string),
		Users:     make(map[string][]string),
		Deleted:   make(map[string]bool),
		Cfg:       cfg,
		Mutex:     &sync.Mutex{},
	}, nil
}

// GetConfig gets config from storage.
func (s *MapStorage) GetConfig() config.Config {
	return s.Cfg
}

// PingDB do nothing.
func (s *MapStorage) PingDB() error {
	return nil
}

// CreateShort creates short url from long.
func (s *MapStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	var err error
	result := make([]string, 0, len(urls))

	s.Lock()
	defer s.Unlock()

	for _, longURL := range urls {
		if _, err := url.ParseRequestURI(longURL); err != nil {
			return nil, fmt.Errorf("wrong url %s", longURL)
		}

		var foundThisURL bool
		newID := fmt.Sprint(len(s.Locations) + 1)

		for id, originalURL := range s.Locations {
			if originalURL == longURL {
				newID = id
				err = ErrExists
				foundThisURL = true
				break
			}
		}
		result = append(result, newID)
		if foundThisURL {
			continue
		}

		s.Locations[newID], s.Users[userID] = longURL, append(s.Users[userID], newID)
	}
	return result, err
}

// GetOriginal gets original url from short.
func (s *MapStorage) GetOriginal(id string) (string, error) {
	var err error

	s.Lock()
	defer s.Unlock()

	if item, ok := s.Locations[id]; ok {
		if s.Deleted[id] {
			err = ErrDeleted
		}
		return item, err
	}
	return "", ErrNotFound
}

// MarkAsDeleted deletes url.
func (s *MapStorage) MarkAsDeleted(userID string, ids ...string) error {
	s.Lock()
	defer s.Unlock()

	for _, id := range ids {
		for _, can := range s.Users[userID] {
			if id == can {
				s.Deleted[id] = true
				break
			}
		}
	}
	return nil
}

// GetURLArrayByUser gets all urls.
func (s *MapStorage) GetURLArrayByUser(userID string) ([]entity.URLs, error) {
	s.Lock()
	defer s.Unlock()

	allShort := s.Users[userID]
	history := make([]entity.URLs, len(allShort))

	for i, id := range allShort {
		original := s.Locations[id]
		history[i] = entity.URLs{
			ShortURL:    fmt.Sprintf("%s/%v", s.Cfg.BaseURL, id),
			OriginalURL: original,
		}
	}
	return history, nil
}

// GetStatistic gets total count of users and urls.
func (s *MapStorage) GetStatistic() (entity.Statistic, error) {
	return entity.Statistic{
		Urls:  len(s.Locations),
		Users: len(s.Users),
	}, nil
}
