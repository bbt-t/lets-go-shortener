package storage

import (
	"io"
	"os"
	"testing"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_GetConfig(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := newFileStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
	assert.NoError(t, err)

	defer os.Remove(cfg.StoragePath)
	defer s.file.Close()
}

func TestFileStorage_PingDB(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := newFileStorage(cfg)
	assert.NoError(t, err)
	assert.NoError(t, s.PingDB())
	assert.NoError(t, err)

	defer os.Remove(cfg.StoragePath)
	defer s.file.Close()
}

func TestFileStorage_GetStatistic(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := newFileStorage(cfg)
	assert.NoError(t, err)

	stat, err := s.GetStatistic()
	assert.NoError(t, err)
	assert.Equal(t, entity.Statistic{
		Urls:  0,
		Users: 0,
	}, stat)

	_, err = s.CreateShort("user12", "https:/123.ru")
	assert.NoError(t, err)
	stat, err = s.GetStatistic()
	assert.NoError(t, err)
	assert.Equal(t, entity.Statistic{
		Urls:  1,
		Users: 0,
	}, stat)
	assert.NoError(t, err)

	defer os.Remove(cfg.StoragePath)
	defer s.file.Close()
}

func TestFileStorage_CreateShort(t *testing.T) {
	cfg := config.GetTestConfig()

	tc := []struct {
		name       string
		urls       []string
		want       []string
		wantInFile []string
		err        error
	}{
		{
			"add single good link to file storage",
			[]string{"https://yandex.ru"},
			[]string{"1"},
			[]string{"https://yandex.ru"},
			nil,
		},
		{
			"add multiple good links to file storage",
			[]string{"https://youtube.com", "https://google.com"},
			[]string{"2", "3"},
			[]string{"https://youtube.com", "https://google.com"},
			nil,
		},
	}

	s, err := newFileStorage(cfg)
	defer os.Remove(cfg.StoragePath)
	defer s.file.Close()
	assert.NoError(t, err)

	for _, test := range tc {
		s, err = newFileStorage(cfg)
		assert.NoError(t, err)

		var res []string
		res, err = s.CreateShort("user12", test.urls...)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.want, res, test.name)
		_, err = s.file.Seek(0, io.SeekStart)
		assert.NoError(t, err)
		var text []byte
		text, err = io.ReadAll(s.file)
		assert.NoError(t, err)

		for _, url := range test.wantInFile {
			assert.Contains(t, string(text), url, test.name)
		}
		s.file.Close()
		os.Remove(cfg.StoragePath)
	}

	assert.NoError(t, s.PingDB())
	assert.NoError(t, err)
}

func TestFileStorage_MarkAsDeleted(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := newFileStorage(cfg)
	defer os.Remove(cfg.StoragePath)
	defer s.file.Close()

	assert.NoError(t, err)

	err = s.MarkAsDeleted("user12", "1234") // do nothing.

	assert.NoError(t, err)
	assert.NoError(t, err)
}
