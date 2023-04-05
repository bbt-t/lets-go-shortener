package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
)

func TestMapStorage_CreateShort(t *testing.T) {
	cfg := config.GetTestConfig()

	tc := []struct {
		name string
		urls []string
		want []string
		loc  map[string]string
		err  error
	}{
		{
			"Add good url",
			[]string{"https://yandex.ru"},
			[]string{"1"},
			map[string]string{"1": "https://yandex.ru"},
			nil,
		},
		{
			"Add good urls",
			[]string{
				"https://yandex.ru",
				"https://google.com",
			},
			[]string{
				"1",
				"2",
			},
			map[string]string{
				"1": "https://yandex.ru",
				"2": "https://google.com",
			},
			nil,
		},
		{
			"Add same urls",
			[]string{
				"https://yandex.ru",
				"https://yandex.ru",
			},
			[]string{"1", "1"},
			map[string]string{"1": "https://yandex.ru"},
			ErrExists,
		},
		{
			"Add bad url",
			[]string{"not_url"},
			nil,
			map[string]string{},
			errors.New("wrong url not_url"),
		},
		{
			"Add bad urls",
			[]string{"not_url", "not_url2"},
			nil,
			map[string]string{},
			errors.New("wrong url not_url"),
		},
		{
			"Don't pass urls",
			nil,
			[]string{},
			map[string]string{},
			nil,
		},
	}

	for _, test := range tc {
		s, err := NewMapStorage(cfg)
		assert.NoError(t, err, test.name)
		res, err := s.CreateShort("user12", test.urls...)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.want, res, test.name)
		assert.Equal(t, test.loc, s.Locations)
	}

}

func TestMapStorage_GetConfig(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
}

func TestMapStorage_GetOriginal(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	tc := []struct {
		name    string
		loc     map[string]string
		deleted map[string]bool
		id      string
		result  string
		err     error
	}{
		{
			"Get long exists url",
			map[string]string{"1": "https://yandex.ru"},
			map[string]bool{"1": false},
			"1",
			"https://yandex.ru",
			nil,
		},
		{
			"Get long non-exists url",
			map[string]string{"1": "https://yandex.ru"},
			map[string]bool{"1": false},
			"2",
			"",
			ErrNotFound,
		},
		{
			"Get long deleted url",
			map[string]string{"1": "https://yandex.ru"},
			map[string]bool{"1": true},
			"1",
			"https://yandex.ru",
			ErrDeleted,
		},
	}

	for _, test := range tc {
		s.Locations = test.loc
		s.Deleted = test.deleted
		res, err := s.GetOriginal(test.id)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.result, res, test.name)
	}
}

func TestMapStorage_MarkAsDeleted(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	tc := []struct {
		name        string
		loc         map[string]string
		users       map[string][]string
		deleted     map[string]bool
		id          string
		wantDeleted map[string]bool
		err         error
	}{
		{
			"user can delete url",
			map[string]string{"1": "https://yandex.ru"},
			map[string][]string{
				"user1": {"1"},
			},
			map[string]bool{"1": false},
			"1",
			map[string]bool{"1": true},
			nil,
		},
		{
			"user can't delete url",
			map[string]string{
				"1": "https://yandex.ru",
				"2": "https://google.com",
			},
			map[string][]string{
				"user1": {"2"},
			},
			map[string]bool{
				"1": false,
				"2": false,
			},
			"1",
			map[string]bool{
				"1": false,
				"2": false,
			},
			nil,
		},
	}

	for _, test := range tc {
		s.Locations = test.loc
		s.Users = test.users
		s.Deleted = test.deleted

		err := s.MarkAsDeleted("user1", test.id)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.wantDeleted, s.Deleted, test.name)
	}

}

func TestMapStorage_GetURLArrayByUser(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	tc := []struct {
		name   string
		loc    map[string]string
		users  map[string][]string
		cookie string
		want   []entity.URLs
		err    error
	}{

		{
			"no history",
			map[string]string{
				"1": "https://yandex.ru",
				"2": "https://google.com",
			},
			map[string][]string{
				"user1": {"1"}, "user2": {"2"},
			},
			"user_unknown",
			[]entity.URLs{},
			nil,
		},
		{
			"get history first user",
			map[string]string{
				"1": "https://yandex.ru",
				"2": "https://google.com",
			},
			map[string][]string{
				"user1": {"1"}, "user2": {"2"},
			},
			"user1",
			[]entity.URLs{
				{
					ShortURL:    cfg.BaseURL + "/1",
					OriginalURL: "https://yandex.ru",
				},
			},
			nil,
		},
		{
			"get history second user",
			map[string]string{
				"1": "https://yandex.ru",
				"2": "https://google.com",
			},
			map[string][]string{
				"user1": {"1"}, "user2": {"2"},
			},
			"user2",
			[]entity.URLs{
				{
					ShortURL:    cfg.BaseURL + "/2",
					OriginalURL: "https://google.com",
				},
			},
			nil,
		},
	}

	for _, test := range tc {
		s.Locations = test.loc
		s.Users = test.users

		res, err := s.GetURLArrayByUser(test.cookie)
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.want, res)
	}

}

func TestMapStorage_PingDB(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	assert.NoError(t, s.PingDB(), "failed ping test")
}
