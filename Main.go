package main

import (
	"forumjs/backend"

	"github.com/joho/godotenv"
)

func main() {
	// load .env
	godotenv.Load()

	// Start server
	backend.Server()
}
