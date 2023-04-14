// Package handlers gets handlers for service endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
	"github.com/bbt-t/lets-go-shortener/internal/usecase"

	"github.com/go-chi/chi/v5"
)

// ShortenerHandler struct for service layer.
type ShortenerHandler struct {
	cfg     config.Config
	storage *usecase.ShortenerService
}

// NewShortenerHandler gets new handlers service.
func NewShortenerHandler(cfg config.Config, s *usecase.ShortenerService) *ShortenerHandler {
	return &ShortenerHandler{
		cfg:     cfg,
		storage: s,
	}
}

// Ping DataBase.
func Ping(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.storage.PingDB(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

// DeleteURL deletes url from storage.
func DeleteURL(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userID := userCookie.Value

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", http.StatusBadRequest)
			return
		}

		var toDelete []string
		if err = json.Unmarshal(resBody, &toDelete); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = s.storage.MarkAsDeleted(userID, toDelete...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// URLBatch shortens batch of urls in single request.
func URLBatch(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqURLs, respURLs []entity.URLBatch

		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userID := userCookie.Value

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(resBody, &reqURLs); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		respURLs, err = ShortURLs(s.storage, userID, reqURLs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(respURLs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	}
}

// RecoverAllURL gets history of your urls.
func RecoverAllURL(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value

		history, err := s.storage.GetURLArrayByUser(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if len(history) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(history)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
}

// RecoverOriginalURL sends person to page, which url was shortened.
func RecoverOriginalURL(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id parameter", http.StatusBadRequest)
			return
		}

		url, err := s.storage.GetOriginal(id)
		if errors.Is(err, storage.ErrDeleted) {
			http.Error(w, "url is deleted", http.StatusGone)
			return
		}
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// RecoverOriginalURLPost creates new short OriginalURL.
func RecoverOriginalURLPost(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqJSON entity.ReqJSON

		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		userID := userCookie.Value

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || string(resBody) == "" {
			http.Error(w, "wrong body", http.StatusBadRequest)
			return
		}
		switch r.Header.Get("Content-Type") {
		case "application/json":
			{
				if err := json.Unmarshal(resBody, &reqJSON); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				res, errCreate := s.storage.CreateShort(userID, reqJSON.URL)
				if errCreate != nil && !errors.Is(errCreate, storage.ErrExists) {
					http.Error(w, errCreate.Error(), http.StatusBadRequest)
					return
				}

				respJSON, err := json.Marshal(entity.RespJSON{Result: s.cfg.BaseURL + "/" + res[0]})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				if errors.Is(errCreate, storage.ErrExists) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusCreated)
				}
				w.Write(respJSON)
			}
		default:
			{
				res, errCreate := s.storage.CreateShort(userID, string(resBody))
				if errCreate != nil && !errors.Is(errCreate, storage.ErrExists) {
					http.Error(w, errCreate.Error(), 400)
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				if errors.Is(errCreate, storage.ErrExists) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusCreated)
				}
				w.Write([]byte(s.cfg.BaseURL + "/" + res[0]))
			}
		}
	}
}

// StatisticHandler returns total urls and users.
func StatisticHandler(s *ShortenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := s.storage.GetStatistic()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(stats)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ShortURLs shorts many urls.
func ShortURLs(s *usecase.ShortenerService, userID string, urlsJSON []entity.URLBatch) ([]entity.URLBatch, error) {
	urls := make([]string, len(urlsJSON))
	resultJSON := make([]entity.URLBatch, len(urlsJSON))

	for i := range urlsJSON {
		urls[i] = urlsJSON[i].OriginalURL
	}

	result, err := s.CreateShort(userID, urls...)
	if err != nil && err != storage.ErrExists {
		return nil, err
	}

	for i, v := range urlsJSON {
		resultJSON[i] = entity.URLBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      s.GetConfig().BaseURL + "/" + result[i],
		}
	}

	return resultJSON, err
}
