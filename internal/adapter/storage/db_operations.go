package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/entity"
)

// PingDB check connection to storage.
func (s *dbStorage) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.db.PingContext(ctx)
}

// CreateShort creates short url from original.
func (s *dbStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	var isErr409 error
	result := make([]string, 0, len(urls))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := s.db.Begin()
	defer tx.Rollback()

	if err != nil {
		return result, err
	}

	stmt, err := tx.PrepareContext(
		ctx,
		"INSERT INTO items (id, url, cookie) VALUES ($1, $2, $3)",
	)
	if err != nil {
		return result, err
	}
	defer stmt.Close()

	for _, url := range urls {
		var isAdded bool

		rows, err := s.db.QueryContext(
			ctx,
			"SELECT id FROM items WHERE url = $1 LIMIT 1",
			url,
		)
		if err != nil {
			return result, err
		}
		for rows.Next() {
			var id string

			err = rows.Scan(&id)
			if err != nil {
				return result, err
			}
			isErr409, isAdded, result = ErrExists, true, append(result, id)
		}

		if err := rows.Err(); err != nil {
			return result, err
		}
		if !isAdded {
			s.lastID++
			newID := fmt.Sprint(s.lastID)
			if _, err := stmt.ExecContext(ctx, newID, url, userID); err != nil {
				return result, err
			}
			result = append(result, newID)
		}
	}

	err = tx.Commit()
	if err != nil {
		return result, err
	}

	return result, isErr409
}

// GetOriginal gets original url from short.
func (s *dbStorage) GetOriginal(id string) (string, error) {
	var (
		original string
		deleted  bool
	)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(
		ctx,
		"SELECT url, deleted FROM items WHERE id=$1 LIMIT 1",
		id,
	)
	err := row.Scan(&original, &deleted)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err := row.Err(); err != nil {
		return "", err
	}
	if deleted {
		return "", ErrDeleted
	}
	return original, nil
}

// MarkAsDeleted deletes url.
func (s *dbStorage) MarkAsDeleted(userID string, ids ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := s.db.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(
		ctx,
		"UPDATE items SET deleted = true WHERE id = $1 AND cookie = $2",
	)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id, userID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetURLArrayByUser gets all urls.
func (s *dbStorage) GetURLArrayByUser(userID string) ([]entity.URLs, error) {
	var history []entity.URLs

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, url FROM items WHERE cookie=$1",
		userID,
	)

	if err != nil {
		return history, err
	}

	defer rows.Close()

	for rows.Next() {
		var id, original string

		err = rows.Scan(&id, &original)

		if err != nil {
			return history, err
		}

		history = append(
			history,
			entity.URLs{
				ShortURL:    id,
				OriginalURL: original,
			},
		)

	}
	if err := rows.Err(); err != nil {
		return history, err
	}

	return history, nil
}

// GetStatistic gets total count of users and urls.
func (s *dbStorage) GetStatistic() (entity.Statistic, error) {
	stat := entity.Statistic{
		Urls: s.lastID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT cookie) FROM items;")

	err := row.Scan(&stat.Users)
	if err != nil {
		return stat, err
	}
	if err := row.Err(); err != nil {
		return stat, err
	}
	return stat, nil
}
