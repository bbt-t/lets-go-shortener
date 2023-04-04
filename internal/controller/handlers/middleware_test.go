package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCookieMiddleware(t *testing.T) {
	var validCookie string

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		assert.NoError(t, ok)
		assert.NotEmpty(t, userID.Value)
		validCookie = userID.Value
	})

	CookieMiddleware(next).ServeHTTP(w, request)

	badCookie, expiration := "badCookie12", time.Now().Add(365*24*time.Hour)

	cookie := http.Cookie{
		Name:    "userID",
		Value:   badCookie,
		Expires: expiration,
		Path:    "/",
	}
	request.AddCookie(&cookie)

	next = func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		assert.NoError(t, ok)
		assert.NotEqual(t, badCookie, userID.Value)
	}

	cookie = http.Cookie{Name: "userID", Value: validCookie, Expires: expiration, Path: "/"}
	request.AddCookie(&cookie)

	next = func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Cookie("userID")
		assert.NoError(t, ok)
		assert.Equal(t, validCookie, userID.Value)
	}
}
