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

// dbStorage is storage that uses DB.
type dbStorage struct {
	Cfg    config.Config
	DB     *sql.DB
	LastID int
}

// GetConfig gets config from storage.
func (s *dbStorage) GetConfig() config.Config {
	return s.Cfg
}

// newDBStorage creates new DB storage.
func newDBStorage(cfg config.Config) (*dbStorage, error) {
	s := &dbStorage{Cfg: cfg}

	db, err := sql.Open("pgx", cfg.BasePath)

	if err != nil {
		return s, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = MigrateUP(db, cfg)

	if err != nil {
		log.Fatalln("Failed migrate DB: ", err)
		return s, err
	}

	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM items")

	err = row.Scan(&s.LastID)

	if err != nil {
		return s, err
	}

	err = row.Err()
	if err != nil {
		return s, err
	}

	s.DB = db

	return s, nil
}

// MigrateUP DB migrations.
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
