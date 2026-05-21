package main

import "forumjs/backend" 


func main() {
	// Création de la base de données
	backend.CreateDatabase()

	// Création des tables
	backend.CreateTables()
	
	// Démarrage du serveur
	backend.Server()
}
