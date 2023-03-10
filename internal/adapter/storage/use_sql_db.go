package storage

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/bbt-t/lets-go-shortener/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// dbStorage is storage that uses db.
type dbStorage struct {
	cfg    config.Config
	db     *sql.DB
	lastID int
}

// GetConfig gets config from storage.
func (s *dbStorage) GetConfig() config.Config {
	return s.cfg
}

// newDBStorage creates new db storage.
func newDBStorage(cfg config.Config) (*dbStorage, error) {
	s := &dbStorage{cfg: cfg}

	db, err := sql.Open("pgx", cfg.BasePath)

	if err != nil {
		return s, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = MigrateUP(db, cfg)

	if err != nil {
		log.Fatalln("Failed migrate db: ", err)
		return s, err
	}

	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM items")

	err = row.Scan(&s.lastID)

	if err != nil {
		return s, err
	}

	err = row.Err()
	if err != nil {
		return s, err
	}

	s.db = db

	return s, nil
}

// MigrateUP db migrations.
func MigrateUP(db *sql.DB, cfg config.Config) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Printf("Failed create postgres instance: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		cfg.DBMigrationPath,
		"pgx",
		driver,
	)

	if err != nil {
		log.Printf("Failed create migration instance: %v\n", err)
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed migrate: ", err)
		return err
	}
	return nil
}
