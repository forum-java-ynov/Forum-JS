package backend

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var dbPath = "database/database.db"

type Comment struct {
	ID       int
	Username string
	Content  string
	Likes    int
	Dislikes int
}

type Post struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Theme     string
	Username  string
	Likes     int
	Dislikes  int
	Comments  []Comment
}

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
		image_path TEXT,
		theme TEXT,
		user_id INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
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
	CREATE TABLE IF NOT EXISTS post_dislike (
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

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS comment_dislike (
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

func getUserIDValue(userID string) (interface{}, error) {
	if id, err := strconv.Atoi(userID); err == nil {
		return id, nil
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var id int
	err = db.QueryRow("SELECT id FROM users WHERE username = ? OR email = ?;", userID, userID).Scan(&id)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func addPost(title, content, imagePath, theme, userID string) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	uid, err := getUserIDValue(userID)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
	INSERT INTO posts (title, content, image_path, theme, user_id) 
	VALUES (?, ?, ?, ?, ?);
	`, title, content, imagePath, theme, uid)
	if err != nil {
		log.Fatal(err)
	}
}

func getPosts() ([]Post, error) {
	db, err := sql.Open("sqlite", dbPath)
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
			posts.theme,
			COALESCE(users.full_name, users.username, 'Utilisateur'),
			COUNT(DISTINCT post_like.id) as likes,
			COUNT(DISTINCT post_dislike.id) as dislikes
		FROM posts
		LEFT JOIN post_like ON posts.id = post_like.post_id
		LEFT JOIN post_dislike ON posts.id = post_dislike.post_id
		LEFT JOIN users ON posts.user_id = users.id
		GROUP BY posts.id;
	`)
	var likes int

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var id int
		var title, content, theme, username string
		var imagePath sql.NullString
		var dislikes int
		if err := rows.Scan(&id, &title, &content, &imagePath, &theme, &username, &likes, &dislikes); err != nil {
			return nil, err
		}
		post := Post{
			ID:       id,
			Title:    title,
			Content:  content,
			Theme:    theme,
			Username: username,
			Likes:    likes,
			Dislikes: dislikes,
		}
		if imagePath.Valid {
			post.ImagePath = imagePath.String
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func addCommente(postID int, userID string, content string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	uid, err := getUserIDValue(userID)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO comments (post_id, user_id, content) 
	VALUES (?, ?, ?);
	`, postID, uid, content)
	return err
}

func getComments(postID string) ([]Comment, error) {
	db, err := sql.Open("sqlite", "database/database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`
	SELECT 
		comments.id,
		COALESCE(users.full_name, users.username, 'Utilisateur'),
		comments.content,
		COUNT(DISTINCT comment_like.id) as likes,
		COUNT(DISTINCT comment_dislike.id) as dislikes
	FROM comments 
	LEFT JOIN users ON comments.user_id = users.id
	LEFT JOIN comment_like ON comments.id = comment_like.comments_id
	LEFT JOIN comment_dislike ON comments.id = comment_dislike.comments_id
	WHERE comments.post_id = ?
	GROUP BY comments.id;
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var id int
		var username, content string
		var likes, dislikes int
		if err := rows.Scan(&id, &username, &content, &likes, &dislikes); err != nil {
			return nil, err
		}
		comment := Comment{
			ID:       id,
			Username: username,
			Content:  content,
			Likes:    likes,
			Dislikes: dislikes,
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
		name, name, email, googleID,
	)
	return err
}

func likepost(postid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO post_like (post_id, user_id) VALUES (?, ?)",
		postid, userID,
	)

	return err
}

func dislikepost(postid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO post_dislike (post_id, user_id) VALUES (?, ?)",
		postid, userID,
	)

	return err
}

func likecomment(commentid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO comment_like (comments_id, user_id) VALUES (?, ?)",
		commentid, userID,
	)

	return err
}

func dislikecomment(commentid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO comment_dislike (comments_id, user_id) VALUES (?, ?)",
		commentid, userID,
	)

	return err
}

func deletelikecomment(commentid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM comment_like WHERE comments_id = ? AND user_id = ?;",
		commentid, userID,
	)

	return err
}

func deletedislikecomment(commentid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM comment_dislike WHERE comments_id = ? AND user_id = ?;",
		commentid, userID,
	)

	return err
}

func deletelikepost(postid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM post_like WHERE post_id = ? AND user_id = ?;",
		postid, userID,
	)

	return err
}

func deletedislikepost(postid string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM post_dislike WHERE post_id = ? AND user_id = ?;",
		postid, userID,
	)

	return err
}

func editcomment(commentID int, content string, userID string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	uid, err := getUserIDValue(userID)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE comments SET content = ? WHERE id = ? AND user_id = ?;",
		content, commentID, uid,
	)

	return err
}