package backend

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var dbPath = "database/database.db"

func CreateDatabase() {
	os.MkdirAll("database", 0755)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connecté à " + dbPath)
}

func CreateTables() {
	db, err := sql.Open("sqlite", dbPath)
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

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		image_path TEXT
	);
	`)

	if err != nil {
		log.Fatal(err)
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func insertUser(fullName, username, email, password, verifPassword string) error {
	if password != verifPassword {
		return fmt.Errorf("passwords do not match")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	hpassword, _ := HashPassword(password)

	_, err = db.Exec(`
    INSERT INTO users (full_name, username, email, password) 
    VALUES (?, ?, ?, ?);
    `, fullName, username, email, hpassword)

	if err != nil {
		if strings.Contains(err.Error(), "users.email") {
			return fmt.Errorf("email déjà utilisé")
		}
		if strings.Contains(err.Error(), "users.username") {
			return fmt.Errorf("username déjà utilisé")
		}
		return err
	}

	return nil
}

func loginUser(username, password string) (bool, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var storedPassword string
	err = db.QueryRow(`SELECT password FROM users WHERE username = ?;`, username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return CheckPasswordHash(password, storedPassword), nil
}

func addPost(title, content, imagePath string) {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
	INSERT INTO posts (title, content, image_path) 
	VALUES (?, ?, ?);
	`, title, content, imagePath)
	if err != nil {
		log.Fatal(err)
	}
}

func getPosts() ([]map[string]string, error) {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT title, content, image_path FROM posts;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []map[string]string
	for rows.Next() {
		var title, content, imagePath string
		if err := rows.Scan(&title, &content, &imagePath); err != nil {
			return nil, err
		}
		post := map[string]string{
			"title":      title,
			"content":    content,
			"image_path": imagePath,
		}
		posts = append(posts, post)
	}
	return posts, nil
}
