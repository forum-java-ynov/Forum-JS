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

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
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
	rows, err := db.Query(`SELECT id, title, content, image_path FROM posts;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []map[string]string
	for rows.Next() {
		var id int
		var title, content string
		var imagePath sql.NullString
		if err := rows.Scan(&id, &title, &content, &imagePath); err != nil {
			return nil, err
		}
		post := map[string]string{
			"id":         fmt.Sprint(id),
			"title":      title,
			"content":    content,
			"image_path": "",
		}
		if imagePath.Valid {
			post["image_path"] = imagePath.String
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func addCommente(postID int, content string) error {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		return err
	}

	defer db.Close()
	_, err = db.Exec(`
	INSERT INTO comments (post_id, user_id, content) 
	VALUES (?, ?, ?);
	`, postID, 1, content)
	return err
}

func getComments(postID string) ([]map[string]string, error) {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT users.username, comments.content 
	FROM comments 
	JOIN users ON comments.user_id = users.id 
	WHERE comments.post_id = ?;`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []map[string]string
	for rows.Next() {
		var username, content string
		if err := rows.Scan(&username, &content); err != nil {
			return nil, err
		}
		comment := map[string]string{
			"username": username,
			"content":  content,
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
