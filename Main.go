package main

import (
	"forumjs/backend"

	"github.com/joho/godotenv"
)

func main() {
	// Charger le .env
	godotenv.Load()
	backend.InitOauth()
	// Création de la base de données
	backend.CreateDatabase()

	// Création des tables
	backend.CreateTables()

	// Démarrage du serveur
	backend.Server()
}
