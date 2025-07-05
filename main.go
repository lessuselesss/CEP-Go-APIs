package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		// Optionally, you can exit here if .env is critical
		// os.Exit(1)
	}

	// This is the main entry point for the application.
	// The test suite targets the library packages, not this executable.

	// Example of accessing an environment variable:
	// circularAddress := os.Getenv("CIRCULAR_ADDRESS")
	// if circularAddress != "" {
	// 	log.Printf("CIRCULAR_ADDRESS: %s", circularAddress)
	// }
}
