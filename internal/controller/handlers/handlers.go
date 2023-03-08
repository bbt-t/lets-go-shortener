// Package handlers gets handlers for service endpoints.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/entity"

	"github.com/go-chi/chi/v5"
)

// Ping DataBase.
func Ping(s storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.PingDB()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

// DeleteURL deletes url from storage.
func DeleteURL(s storage.Repository) http.HandlerFunc {
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
		err = json.Unmarshal(resBody, &toDelete)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.MarkAsDeleted(userID, toDelete...)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// URLBatch shortens batch of urls in single request.
func URLBatch(s storage.Repository) http.HandlerFunc {
	cfg := s.GetConfig()
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

		var reqURLs, respURLs []entity.URLBatch

		err = json.Unmarshal(resBody, &reqURLs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var urls []string

		for _, v := range reqURLs {
			urls = append(urls, v.OriginalURL)
		}

		res, err := s.CreateShort(userID, urls...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for i, v := range reqURLs {
			respURLs = append(respURLs, entity.URLBatch{CorrelationID: v.CorrelationID, ShortURL: cfg.BaseURL + "/" + res[i]})
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
func RecoverAllURL(s storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := r.Cookie("userID")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		userID := userCookie.Value

		history, err := s.GetURLArrayByUser(userID)
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
func RecoverOriginalURL(s storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id parameter", http.StatusBadRequest)
			return
		}
		url, err := s.GetOriginal(id)

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
func RecoverOriginalURLPost(s storage.Repository) http.HandlerFunc {
	cfg := s.GetConfig()
	return func(w http.ResponseWriter, r *http.Request) {
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
				var reqJSON entity.ReqJSON

				err := json.Unmarshal(resBody, &reqJSON)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				res, err2 := s.CreateShort(userID, reqJSON.URL)

				if err2 != nil && !errors.Is(err2, storage.ErrExists) {
					http.Error(w, err2.Error(), http.StatusBadRequest)
					return
				}

				respJSON, err := json.Marshal(entity.RespJSON{Result: cfg.BaseURL + "/" + res[0]})

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				if errors.Is(err2, storage.ErrExists) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusCreated)
				}
				w.Write(respJSON)
			}
		default:
			{
				res, err2 := s.CreateShort(userID, string(resBody))
				if err2 != nil && !errors.Is(err2, storage.ErrExists) {
					http.Error(w, err2.Error(), 400)
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				if errors.Is(err2, storage.ErrExists) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusCreated)
				}
				w.Write([]byte(cfg.BaseURL + "/" + res[0]))
			}

		}

	}
}
