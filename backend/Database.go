package backend

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"fmt"

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

func CreateTables() {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func insertUser(fullName, username, email, password, verifPassword string) error {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		return err
	}
	if password != verifPassword {
		return fmt.Errorf("passwords do not match")
	}
	defer db.Close()
	_, err = db.Exec(`
	INSERT INTO users (full_name, username, email, password) 
	VALUES (?, ?, ?, ?);
	`, fullName, username, email, password)

	if err != nil {
		if strings.Contains(err.Error(), "users.email") {
			return fmt.Errorf("email déjà utilisé")
		}

		if strings.Contains(err.Error(), "users.username") {
			return fmt.Errorf("username déjà utilisé")
		}

		return err
	}

	return err
}

func loginUser(username, password string) (bool, error) {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		return false, err
	}
	defer db.Close()

	var storedPassword string
	err = db.QueryRow(`
	SELECT password FROM users WHERE username = ?;
	`, username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return storedPassword == password, nil
}
