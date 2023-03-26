package main

import (
	"log"
	"os"
)

func main() {
	// want "you shouldn't use os.Exit in function main"
	log.Println("Hello, world!")
	os.Exit(0)
}
