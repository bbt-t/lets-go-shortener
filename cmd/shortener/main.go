package main

import (
	"log"

	"github.com/bbt-t/lets-go-shortener/internal/app"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	_ "github.com/bbt-t/lets-go-shortener/pkg/logging"
)

// Run example:  go run -ldflags "-X main.buildVersion=0.19 -X main.buildDate=12.03.23 -X main.buildCommit=iter19" cmd/shortener/main.go.

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	showBuildInfo()
	app.Run(config.GetConfig())
}

func showBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	log.Println("Build version:", buildVersion)
	log.Println("Build date:", buildDate)
	log.Println("Build commit:", buildCommit)
}
