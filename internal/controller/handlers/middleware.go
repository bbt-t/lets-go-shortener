package handlers

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// compress response.

// secretKey is key for cookie auth.
var secretKey = []byte("super secret key")

// [32 bytes signature][8 bytes userID].

// gzipWriter struct for sending gzip packed response.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write - sends gzip packed response.
func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer responsible for compression
	return w.Writer.Write(b)
}

// generateRandom generates random bytes for cookie authentication.
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// CookieMiddleware check if user is authorized.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		if ok == nil {
			id, err := hex.DecodeString(userID.Value)
			if err != nil || len(id) != 40 {
				http.Error(w, err.Error(), 400)
				return
			}
			signSrc := id[:32]
			id = id[32:]
			h := hmac.New(sha256.New, secretKey)
			h.Write(id)
			sign := h.Sum(nil)
			if !hmac.Equal(signSrc, sign) {
				log.Println("failed to verify")
				ok = errors.New("failed to verify")
			}
		}
		if ok != nil {
			randomID, err := generateRandom(8)
			h := hmac.New(sha256.New, secretKey)
			h.Write(randomID)
			sign := h.Sum(nil)
			if err != nil {
				http.Error(w, err.Error(), 400)
			}

			cookie := http.Cookie{
				Name:    "userID",
				Value:   hex.EncodeToString(append(sign, randomID...)),
				Expires: time.Now().Add(24 * time.Hour),
				Path:    "/",
			}
			http.SetCookie(w, &cookie)
			r.AddCookie(&cookie)
		}
		next.ServeHTTP(w, r)
	})
}

// GzipRequest accepts gzip request.
func GzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		// переменная r будет читать входящие данные и распаковывать их.
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Println(err)
			return
		}
		defer reader.Close()

		r.Body = reader

		next.ServeHTTP(w, r)
	})
}

// GzipHandle sends gzip packed data.
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(
			gzipWriter{ResponseWriter: w, Writer: gz},
			r,
		)
	})
}
