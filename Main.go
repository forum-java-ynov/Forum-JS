package main

import "forumjs/backend" 


func main() {
	// Création de la base de données
	backend.CreateDatabase()
	// Démarrage du serveur
	backend.Server()
}
