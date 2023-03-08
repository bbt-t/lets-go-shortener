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

// FileStorage struct of file storage.
type FileStorage struct {
	Cfg    config.Config
	File   *os.File
	LastID int
	*sync.Mutex
}

// GetConfig gets config.
func (s *FileStorage) GetConfig() config.Config {
	return s.Cfg
}

// PingDB does nothing.
func (s *FileStorage) PingDB() error {
	return nil
}

// NewFileStorage creates new file storage.
func NewFileStorage(cfg config.Config) (*FileStorage, error) {
	s := &FileStorage{Cfg: cfg, Mutex: &sync.Mutex{}}

	if cfg.StoragePath == "" {
		return s, errors.New("empty file path")
	}

	file, err := os.OpenFile(cfg.StoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0777)
	if err != nil {
		return s, err
	}

	s.File = file

	scanner := bufio.NewScanner(file)
	id := 0
	for scanner.Scan() {
		id++
	}

	if err := scanner.Err(); err != nil {
		return s, err
	}

	s.LastID = id

	return s, nil
}

// CreateShort creates short url from original.
func (s *FileStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	var builder strings.Builder

	s.Lock()
	defer s.Unlock()

	s.File.Seek(2, io.SeekEnd)

	result := make([]string, 0, len(urls))

	for _, original := range urls {
		builder.WriteString(original)
		builder.WriteRune('\n')
		s.LastID++
		result = append(result, fmt.Sprint(s.LastID))
	}

	_, err := s.File.Write([]byte(builder.String()))
	if err != nil {
		return nil, err
	}
	err = s.File.Sync()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetOriginal gets original url from short.
func (s *FileStorage) GetOriginal(id string) (string, error) {
	s.Lock()
	defer s.Unlock()

	s.File.Seek(0, io.SeekStart)

	scanner := bufio.NewScanner(s.File)
	i := 0
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
func (s *FileStorage) MarkAsDeleted(userID string, ids ...string) error {
	// do nothing for file storage.
	return nil
}

// GetURLArrayByUser gets history of urls.
func (s *FileStorage) GetURLArrayByUser(_ string) ([]entity.URLs, error) {
	var allURLs []entity.URLs

	scanner := bufio.NewScanner(s.File)
	id := 0
	for scanner.Scan() {
		id++
		original := scanner.Text()
		allURLs = append(
			allURLs, entity.URLs{
				ShortURL:    fmt.Sprintf("%s/%v", s.Cfg.BaseURL, id),
				OriginalURL: original,
			},
		)
	}

	return allURLs, scanner.Err()
}
