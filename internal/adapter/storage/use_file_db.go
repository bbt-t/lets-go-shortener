package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
)

// fileStorage struct of file storage.
type fileStorage struct {
	cfg    config.Config
	file   *os.File
	lastID int
	*sync.Mutex
}

// GetConfig gets config.
func (s *fileStorage) GetConfig() config.Config {
	return s.cfg
}

// PingDB does nothing.
func (s *fileStorage) PingDB() error {
	return nil
}

// newFileStorage creates new file storage.
func newFileStorage(cfg config.Config) (*fileStorage, error) {
	var id int

	s := &fileStorage{cfg: cfg, Mutex: &sync.Mutex{}}

	if cfg.StoragePath == "" {
		return s, errors.New("empty file path")
	}
	file, err := os.OpenFile(
		cfg.StoragePath,
		os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC,
		0700,
	)
	if err != nil {
		return s, err
	}

	s.file = file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id++
	}
	if err := scanner.Err(); err != nil {
		return s, err
	}

	s.lastID = id

	return s, nil
}

// CreateShort creates short url from original.
func (s *fileStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	var builder strings.Builder

	s.Lock()
	defer s.Unlock()

	s.file.Seek(2, io.SeekEnd)

	result := make([]string, 0, len(urls))

	for _, original := range urls {
		builder.WriteString(original)
		builder.WriteRune('\n')
		s.lastID++
		result = append(result, fmt.Sprint(s.lastID))
	}

	_, err := s.file.Write([]byte(builder.String()))
	if err != nil {
		return nil, err
	}
	err = s.file.Sync()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetOriginal gets original url from short.
func (s *fileStorage) GetOriginal(id string) (string, error) {
	var i int

	s.Lock()
	defer s.Unlock()

	s.file.Seek(0, io.SeekStart)

	scanner := bufio.NewScanner(s.file)

	for scanner.Scan() {
		i++
		original := scanner.Text()
		if fmt.Sprint(i) == id {
			return original, scanner.Err()
		}
	}
	return "", ErrNotFound
}

// MarkAsDeleted does nothing.
func (s *fileStorage) MarkAsDeleted(userID string, ids ...string) error {
	// do nothing for file storage.
	return nil
}

// GetURLArrayByUser gets history of urls.
func (s *fileStorage) GetURLArrayByUser(_ string) ([]entity.URLs, error) {
	var (
		id      int
		allURLs []entity.URLs
	)

	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		id++
		original := scanner.Text()
		allURLs = append(
			allURLs, entity.URLs{
				ShortURL:    fmt.Sprintf("%s/%v", s.cfg.BaseURL, id),
				OriginalURL: original,
			},
		)
	}
	return allURLs, scanner.Err()
}
