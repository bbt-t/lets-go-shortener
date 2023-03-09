package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"
)

func ExampleRecoverOriginalURLPost() {
	// data.
	body := "https://yandex.ru"
	// Generating handler.
	cfg := config.GetDefaultConfig()
	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	h := RecoverOriginalURLPost(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()

	cookie := http.Cookie{
		Name:    "userID",
		Value:   "user12cookie",
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	log.Printf("Short OriginalURL is: %s\n", resBody)
	log.Printf("Now in storage: %s\n", s.Locations)
}

func ExampleURLBatch() {
	// data.
	body := `[{"correlation_id": "1", "original_url": "https://yandex.ru"}, 
			  {"correlation_id": "2", "original_url": "https://google.com"}, 
			  {"correlation_id": "3", "original_url": "https://youtube.com"}]`
	// Generating handler.
	cfg := config.GetDefaultConfig()
	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	h := URLBatch(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cookie := http.Cookie{
		Name:    "userID",
		Value:   "user12cookie",
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	log.Printf("Short OriginalURL is: %s\n", resBody)
	log.Printf("Now in storage: %s\n", s.Locations)
}

func ExampleRecoverAllURL() {
	// data.
	userID := "user12cookie"
	// generating storage.
	cfg := config.GetDefaultConfig()
	s, err := storage.NewMapStorage(cfg)

	if err != nil {
		log.Fatal("Failed get storage")
	}

	// creating short urls.
	_, err = s.CreateShort(userID, "https://yandex.ru")

	if err != nil {
		log.Fatal("Failed shorten OriginalURL")
	}

	_, err = s.CreateShort(userID, "https://google.com")

	if err != nil {
		log.Fatal("Failed shorten OriginalURL")
	}
	// Generating handler.
	h := RecoverAllURL(s)

	// Generating request.
	request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()

	cookie := http.Cookie{
		Name:    "userID",
		Value:   userID,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
	request.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, request)

	// Checking result.
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	log.Printf("URLs history is: %s\n", resBody)
}

func ExampleDeleteURL() {
	// data.
	data, userID := `["1"]`, "user12cookie"
	// generating storage.
	s, err := storage.NewMapStorage(config.GetDefaultConfig())
	if err != nil {
		log.Fatal("Failed get storage")
	}
	// creating short urls.
	if _, err = s.CreateShort(userID, "https://ya.ru"); err != nil {
		log.Fatal("Failed shorten OriginalURL")
	}
	if _, err = s.CreateShort(userID, "https://www.python.org"); err != nil {
		log.Fatal("Failed shorten OriginalURL")
	}
	// Generating handler.
	h := DeleteURL(s)

	// Generating request.
	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/user/urls",
		strings.NewReader(data),
	)
	w := httptest.NewRecorder()

	cookie := http.Cookie{
		Name:    "userID",
		Value:   userID,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
	req.AddCookie(&cookie)

	// Serving request using handler.
	h.ServeHTTP(w, req)
}
