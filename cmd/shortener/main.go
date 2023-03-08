package main

import (
	"github.com/bbt-t/lets-go-shortener/internal/app"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	_ "github.com/bbt-t/lets-go-shortener/pkg/logging"
)

func main() {
	app.Run(config.GetConfig())
}
