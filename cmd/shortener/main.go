package main

import (
	"github.com/bbt-t/lets-go-shortener/internal/app"
	"github.com/bbt-t/lets-go-shortener/internal/config"
)

func main() {
	app.Run(config.GetConfig())
}
