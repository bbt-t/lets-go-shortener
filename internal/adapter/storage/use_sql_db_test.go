package storage

import (
	"errors"
	"testing"

	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewDBStorage(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := newDBStorage(cfg)

	assert.NoError(t, err)

	// changing DB to mock DB.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err, "Create new mock DB storage.")
	defer db.Close()

	s.db = db
	_ = mock
}

func TestDBStorage_GetConfig(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := newDBStorage(cfg)

	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
}

func TestDBStorage(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := newDBStorage(cfg)

	assert.NoError(t, err)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual), sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err, "Create new mock DB storage.")
	defer db.Close()

	s.db = db

	// add new links.
	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)").
		WithArgs("1", "https://yandex.ru", "user12", false).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://google.com").
		WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)").
		WithArgs("2", "https://google.com", "user12", false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	result, err := s.CreateShort("user12", "https://yandex.ru", "https://google.com")
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, result)

	assert.NoError(t, mock.ExpectationsWereMet())

	// add existed link.

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectCommit()

	result, err = s.CreateShort("user12", "https://yandex.ru")
	assert.Equal(t, ErrExists, err)
	assert.Equal(t, []string{"1"}, result)

	assert.NoError(t, mock.ExpectationsWereMet())

	// expecting error and rollback.
	ErrRow := errors.New("row error")

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1").RowError(0, ErrRow))

	mock.ExpectRollback()

	_, err = s.CreateShort("user12", "https://yandex.ru")
	assert.Equal(t, ErrRow, err)

	assert.NoError(t, mock.ExpectationsWereMet())

	// ping DB (no error).

	mock.ExpectPing()
	err = s.PingDB()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// ping DB (with error).

	ErrPing := errors.New("failed ping DB")
	mock.ExpectPing().WillReturnError(ErrPing)
	err = s.PingDB()
	assert.Equal(t, ErrPing, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// get long url.

	mock.ExpectQuery("SELECT url, deleted FROM links WHERE id=$1 LIMIT 1").WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"url", "deleted"}).AddRow("https://yandex.ru", false))

	longURL, err := s.GetOriginal("1")
	assert.NoError(t, err)
	assert.Equal(t, "https://yandex.ru", longURL)
	assert.NoError(t, mock.ExpectationsWereMet())

	// get non-existed long url.

	mock.ExpectQuery("SELECT url, deleted FROM links WHERE id=$1 LIMIT 1").WithArgs("3").
		WillReturnRows(sqlmock.NewRows([]string{"url", "deleted"}))

	_, err = s.GetOriginal("3")
	assert.Equal(t, ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// get long url (with error).

	mock.ExpectQuery("SELECT url, deleted FROM links WHERE id=$1 LIMIT 1").WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"url", "deleted"}).AddRow("https://yandex.ru", false).RowError(0, ErrRow))

	_, err = s.GetOriginal("1")

	assert.Equal(t, ErrRow, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// get deleted long url.

	mock.ExpectQuery("SELECT url, deleted FROM links WHERE id=$1 LIMIT 1").WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"url", "deleted"}).AddRow("https://yandex.ru", true))

	_, err = s.GetOriginal("1")
	assert.Equal(t, ErrDeleted, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// get urls history from exists user.

	mock.ExpectQuery("SELECT id, url FROM links WHERE cookie=$1").WithArgs("user12").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow("1", "https://yandex.ru").AddRow("2", "https://google.com"))

	history, err := s.GetURLArrayByUser("user12")

	assert.NoError(t, err)

	assert.Equal(t, history, []entity.URLs{
		{
			OriginalURL: "https://yandex.ru",
			ShortURL:    "1",
		},
		{
			OriginalURL: "https://google.com",
			ShortURL:    "2",
		},
	})

	assert.NoError(t, mock.ExpectationsWereMet())

	// get urls history from non-exists user.

	mock.ExpectQuery("SELECT id, url FROM links WHERE cookie=$1").WithArgs("unknown").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}))

	history, err = s.GetURLArrayByUser("unknown")

	assert.NoError(t, err)
	assert.Equal(t, history, []entity.URLs(nil))
	assert.NoError(t, mock.ExpectationsWereMet())

	// get urls history with error.

	mock.ExpectQuery("SELECT id, url FROM links WHERE cookie=$1").WithArgs("user12").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow("1", "https://yandex.ru").RowError(0, ErrRow))

	history, err = s.GetURLArrayByUser("user12")

	assert.Equal(t, ErrRow, err)
	assert.Equal(t, history, []entity.URLs(nil))
	assert.NoError(t, mock.ExpectationsWereMet())

	// delete url from existed user.
	mock.ExpectBegin()
	mock.ExpectPrepare("UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2")
	mock.ExpectExec("UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2").
		WithArgs("1", "user12").WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err = s.MarkAsDeleted("user12", "1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// delete url from non-existed user.
	mock.ExpectBegin()
	mock.ExpectPrepare("UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2")
	mock.ExpectExec("UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2").
		WithArgs("1", "unknown").WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()

	err = s.MarkAsDeleted("unknown", "1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// delete with error.
	mock.ExpectBegin()
	mock.ExpectPrepare("UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2")
	mock.ExpectExec("UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2").
		WithArgs("1", "unknown").WillReturnError(ErrRow)

	mock.ExpectRollback()

	err = s.MarkAsDeleted("unknown", "1")
	assert.Equal(t, ErrRow, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// get statistic.
	mock.ExpectQuery("SELECT COUNT(DISTINCT cookie) FROM links;").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))

	stat, err := s.GetStatistic()
	assert.NoError(t, err)
	assert.Equal(t, entity.Statistic{
		Urls:  s.lastID,
		Users: 20,
	}, stat)

	assert.NoError(t, mock.ExpectationsWereMet())

}
