package backend

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func CreateDatabase() {
	os.MkdirAll("database", 0755)

	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connecté à database/database.db")
}
