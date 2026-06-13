package main

import (
	"forumjs/backend"

	"github.com/joho/godotenv"
)

func main() {
	// Charger le .env
	godotenv.Load()

	// Démarrage du serveur
	backend.Server()
}
