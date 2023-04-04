package storage

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/bbt-t/lets-go-shortener/internal/config"
)

func BenchmarkDBStorage(b *testing.B) {
	cfg := config.GetConfig()
	cfg.ChangeByPriority(config.GetBenchConfig())
	if cfg.BasePath == "" {
		log.Println("DB path is empty, skipping benchmark.")
		return
	}

	if cfg.StoragePath == "" {
		return
	}

	cfg.DBMigrationPath = "file://../../migrations"
	s, err := newDBStorage(cfg)
	if err != nil {
		log.Fatalln("Failed get storage: ", err)
	}

	b.ResetTimer()
	b.Run("Add new urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			url := fmt.Sprintf(
				"https://random%v/random%v",
				rand.Intn(50000),
				rand.Intn(20000),
			)
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()
			if _, err := s.CreateShort(userID, url); err != nil {
				log.Println(err)
			}
		}
	})

	b.Run("Get original urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(rand.Intn(s.LastID))
			b.StartTimer()

			if _, err := s.GetOriginal(id); err != nil {
				log.Println(err)
			}
		}
	})

	b.Run("Delete urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(rand.Intn(s.LastID))
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			if err := s.MarkAsDeleted(userID, id); err != nil {
				log.Println(err)
			}
		}
	})

	b.Run("Get history", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			if _, err := s.GetURLArrayByUser(userID); err != nil {
				log.Println(err)
			}
		}
	})

}

func BenchmarkMapStorage(b *testing.B) {
	cfg := config.GetConfig()
	s, err := NewMapStorage(cfg)
	if err != nil {
		log.Fatalln("Failed get storage: ", err)
	}

	b.ResetTimer()
	b.Run("Add new urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			url := fmt.Sprintf(
				"https://random%v/random%v",
				rand.Intn(5000),
				rand.Intn(2000),
			)
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			if _, err := s.CreateShort(userID, url); err != nil {
				log.Println(err)
			}
		}
	})

	b.Run("Get original urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(rand.Intn(len(s.Locations)))
			b.StartTimer()

			if _, err := s.GetOriginal(id); err != nil {
				log.Println(err)
			}
		}
	})

	b.Run("Delete urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(len(s.Locations))
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			if err := s.MarkAsDeleted(userID, id); err != nil {
				log.Println(err)
			}
		}
	})

	b.Run("Get history", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			if _, err := s.GetURLArrayByUser(userID); err != nil {
				log.Println(err)
			}
		}
	})

}
