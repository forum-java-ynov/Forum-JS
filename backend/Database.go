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
		password TEXT,
	    google_id TEXT
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
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS post_like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`)

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comment_like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		comments_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		FOREIGN KEY (comments_id) REFERENCES comments(id),
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
	rows, err := db.Query(`
		SELECT 
			posts.id,
			posts.title,
			posts.content,
			posts.image_path,
			COUNT(post_like.id) as likes
		FROM posts
		LEFT JOIN post_like ON posts.id = post_like.post_id
		GROUP BY posts.id;
	`)
	var likes int

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []map[string]string
	for rows.Next() {
		var id int
		var title, content string
		var imagePath sql.NullString
		if err := rows.Scan(&id, &title, &content, &imagePath, &likes); err != nil {
			return nil, err
		}
		post := map[string]string{
			"id":         fmt.Sprint(id),
			"title":      title,
			"content":    content,
			"image_path": "",
			"likes": fmt.Sprint(likes),
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

func deletePost(id int) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
	DELETE FROM posts WHERE id = ?;`, id)
	return err
}

// google auth
func userExistsByEmail(email string) (bool, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

func insertGoogleUser(name, email, googleID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO users (full_name, username, email, google_id) VALUES (?, ?, ?, ?)",
		name, email, email, googleID,
	)
	return err
}


func likepost(postid string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO post_like (post_id, user_id) VALUES (?, ?)",
		postid, 1,
	)

	return err
}

func likecomment(commentid string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO comment_like (comment_id, user_id) VALUES (?, ?)",
		commentid, 1,
	)

	return err
}

func deletelikecomment(commentid string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM comment_like WHERE comment_id = ? AND user_id = ?;",
		commentid, 1,
	)

	return err
}


func deletelikepost(postid string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM post_like WHERE post_id = ? AND user_id = ?;",
		postid, 1,
	)

	return err
}

