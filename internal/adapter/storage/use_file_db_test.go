package storage

import (
	"github.com/bbt-t/lets-go-shortener/internal/entity"
	"testing"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFileStorage_GetConfig(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewStorage(cfg)

	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
	assert.NoError(t, err)
}

func TestFileStorage_PingDB(t *testing.T) {
	s, err := NewStorage(config.GetTestConfig())

	assert.NoError(t, err)
	assert.NoError(t, s.PingDB())
	assert.NoError(t, err)
}

func TestFileStorage_MarkAsDeleted(t *testing.T) {
	s, err := newFileStorage(config.GetTestConfig())

	assert.NoError(t, err)

	err = s.MarkAsDeleted("user12", "1234") // do nothing.

	assert.NoError(t, err)
	assert.NoError(t, err)
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
}
