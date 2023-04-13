package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
	"github.com/bbt-t/lets-go-shortener/internal/usecase"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestURLPostHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		urls     *storage.MapStorage
		error    bool
	}
	cases := []struct {
		name string
		urls *storage.MapStorage
		url  string
		want want
	}{
		{
			"add new url storage",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{},
			},
			"https://google.com",
			want{
				201,
				"2",
				&storage.MapStorage{
					Locations: map[string]string{
						"1": "https://123.ru",
						"2": "https://google.com",
					},
					Mutex: &sync.Mutex{},
					Users: map[string][]string{"123456": {"2"}}},
				false,
			},
		},
		{
			"add bad url to storage",
			&storage.MapStorage{Locations: map[string]string{"1": "https://123.ru"},
				Mutex: &sync.Mutex{},
				Users: map[string][]string{"123456": {"2"}},
			},
			"efjwejfekw",
			want{
				400,
				"wrong url",
				&storage.MapStorage{
					Locations: map[string]string{"1": "https://123.ru"},
					Mutex:     &sync.Mutex{},
					Users:     map[string][]string{"123456": {"2"}},
				},
				true,
			},
		},
		{
			"don't send body",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{"123456": {"2"}},
			},
			"",
			want{
				400,
				"wrong body\n",
				&storage.MapStorage{Locations: map[string]string{"1": "https://123.ru"},
					Mutex: &sync.Mutex{},
					Users: map[string][]string{"123456": {"2"}}},
				true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.url))

			w := httptest.NewRecorder()
			tc.urls.Cfg = config.GetDefaultConfig()

			service := usecase.NewShortenerService(tc.urls.Cfg, tc.urls)
			handlers := NewShortenerHandler(tc.urls.Cfg, service)
			h := RecoverOriginalURLPost(handlers)

			cookie := http.Cookie{
				Name:    "userID",
				Value:   "123456",
				Expires: time.Now().Add(24 * time.Hour),
				Path:    "/",
			}
			request.AddCookie(&cookie)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			if tc.want.error {
				assert.Contains(t, string(resBody), tc.want.response)
			}
			assert.Equal(t, tc.want.urls.Locations, tc.urls.Locations)

		})
	}

}

func TestURLPostJSONHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		urls     *storage.MapStorage
		error    bool
	}
	cases := []struct {
		name string
		urls *storage.MapStorage
		url  string
		want want
	}{
		{
			"add new url storage",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{},
			},
			`{"url":"https://google.com"}`,
			want{
				201,
				"2",
				&storage.MapStorage{
					Locations: map[string]string{"1": "https://123.ru", "2": "https://google.com"},
					Mutex:     &sync.Mutex{},
					Users:     map[string][]string{"123456": {"2"}},
				},
				false,
			},
		},
		{
			"add bad url to storage",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{"123456": {"2"}},
			},
			"{efjwejfekw",
			want{
				400,
				"wrong url\n",
				&storage.MapStorage{
					Locations: map[string]string{"1": "https://123.ru"},
					Mutex:     &sync.Mutex{},
					Users:     map[string][]string{"123456": {"2"}},
				},
				true,
			},
		},
		{
			"don't send body",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{"123456": {"2"}},
			},
			"",
			want{
				400,
				"wrong body\n",
				&storage.MapStorage{
					Locations: map[string]string{"1": "https://123.ru"},
					Mutex:     &sync.Mutex{},
					Users:     map[string][]string{"123456": {"2"}},
				},
				true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tc.url))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			tc.urls.Cfg = config.GetDefaultConfig()

			service := usecase.NewShortenerService(tc.urls.Cfg, tc.urls)
			handlers := NewShortenerHandler(tc.urls.Cfg, service)
			h := RecoverOriginalURLPost(handlers)

			cookie := http.Cookie{
				Name:    "userID",
				Value:   "123456",
				Expires: time.Now().Add(24 * time.Hour),
				Path:    "/",
			}
			request.AddCookie(&cookie)
			h.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			_, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.urls.Locations, tc.urls.Locations)

		})
	}

}

func TestURLGetHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		error    bool
	}
	cases := []struct {
		name string
		urls *storage.MapStorage
		id   string
		want want
	}{
		{
			"get url which in storage",
			&storage.MapStorage{
				Locations: map[string]string{"1": "http://123.ru"},
				Mutex:     &sync.Mutex{},
			},
			"1",
			want{307, "", false},
		},
		{
			"get url which NOT in storage",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
			},
			"2",
			want{404, "not found\n", true},
		},
		{
			"don't send ID parameter",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
			},
			"",
			want{400, "missing id parameter\n", true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			tc.urls.Cfg = config.GetDefaultConfig()

			service := usecase.NewShortenerService(tc.urls.Cfg, tc.urls)
			handlers := NewShortenerHandler(tc.urls.Cfg, service)
			h := RecoverOriginalURL(handlers)

			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tc.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.NoError(t, err)
			if tc.want.error {
				assert.Equal(t, tc.want.response, string(resBody))
			}

		})
	}

}

func TestNewShortURL(t *testing.T) {
	type want struct {
		urls  *storage.MapStorage
		id    string
		error error
	}
	cases := []struct {
		name string
		urls *storage.MapStorage
		url  string
		want want
	}{
		{
			"add new url",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{},
			},
			"https://google.com",
			want{
				&storage.MapStorage{
					Locations: map[string]string{"1": "https://123.ru", "2": "https://google.com"},
					Mutex:     &sync.Mutex{},
					Users:     map[string][]string{},
				},
				"2",
				nil,
			},
		},
		{
			"add bad url",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
				Users:     map[string][]string{},
			},
			"njkjnekjre",
			want{
				&storage.MapStorage{
					Locations: map[string]string{"1": "https://123.ru"},
					Mutex:     &sync.Mutex{},
					Users:     map[string][]string{},
				},
				"",
				errors.New("wrong url"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.urls.CreateShort("123456", tc.url)
			assert.Equal(t, tc.want.urls.Locations, tc.urls.Locations)
			if tc.want.error != nil {
				assert.Contains(t, err.Error(), tc.want.error.Error())
			} else {
				assert.Equal(t, tc.want.id, id[0])
			}
		})
	}
}

func TestGetFullURL(t *testing.T) {
	type want struct {
		url   string
		error error
	}
	cases := []struct {
		name string
		urls *storage.MapStorage
		id   string
		want want
	}{
		{
			"get existed url",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
			},
			"1",
			want{
				"https://123.ru",
				nil,
			},
		},
		{
			"get non-existed url",
			&storage.MapStorage{
				Locations: map[string]string{"1": "https://123.ru"},
				Mutex:     &sync.Mutex{},
			},
			"2",
			want{
				"",
				errors.New("not found"),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := tc.urls.GetOriginal(tc.id)
			if tc.want.error != nil {
				assert.Equal(t, tc.want.error, err)
			} else {
				assert.Equal(t, tc.want.url, url)
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	cfg := config.GetDefaultConfig()
	s, err := storage.NewMapStorage(cfg)
	assert.NoError(t, err)

	service := usecase.NewShortenerService(cfg, s)
	handlers := NewShortenerHandler(cfg, service)

	h := Ping(handlers)

	cookie := http.Cookie{
		Name:    "userID",
		Value:   "user12",
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Path:    "/",
	}
	request.AddCookie(&cookie)
	h.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestService_GetStatistic(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/internal/stats", nil)
	w := httptest.NewRecorder()
	cfg := config.GetTestConfig()
	s, err := storage.NewMapStorage(cfg)

	assert.NoError(t, err)
	assert.NoError(t, err)

	service := usecase.NewShortenerService(cfg, s)
	handlers := NewShortenerHandler(cfg, service)

	h := StatisticHandler(handlers)
	h.ServeHTTP(w, request)

	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	defer result.Body.Close()

	check := entity.Statistic{}

	bytes, err := io.ReadAll(result.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(bytes, &check)
	assert.NoError(t, err)
	assert.Equal(t, entity.Statistic{
		Urls:  0,
		Users: 0,
	}, check)

	// adding url to storage.
	_, err = s.CreateShort("user12", "https://yandex.ru")
	assert.NoError(t, err)

	w = httptest.NewRecorder()

	h.ServeHTTP(w, request)

	result = w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	defer result.Body.Close()

	check = entity.Statistic{}

	bytes, err = io.ReadAll(result.Body)
	assert.NoError(t, err)

	t.Log(string(bytes))

	err = json.Unmarshal(bytes, &check)
	assert.NoError(t, err)
	assert.Equal(t, entity.Statistic{
		Urls:  1,
		Users: 1,
	}, check)

}
