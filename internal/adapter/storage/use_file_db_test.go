package storage

import (
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
	cfg := config.GetTestConfig()
	s, err := NewStorage(cfg)
	assert.NoError(t, err)
	assert.NoError(t, s.PingDB())
	assert.NoError(t, err)
}
