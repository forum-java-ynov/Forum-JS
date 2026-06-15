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

var DB *sql.DB

type Comment struct {
	ID           int
	Username     string
	Content      string
	Likes        int
	Dislikes     int
	UserLiked    bool
	UserDisliked bool
}

type Post struct {
	ID           int
	Title        string
	Content      string
	ImagePath    string
	Theme        string
	Username     string
	Likes        int
	Dislikes     int
	UserLiked    bool
	UserDisliked bool
	Comments     []Comment
}

// InitDB int sqlite connectionand create necesarry tables
func InitDB() {
	os.MkdirAll("database", 0755)

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Impossible d'ouvrir la base de données: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("La base de données ne répond pas: %v", err)
	}

	log.Println("Connecté à " + dbPath)
	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name TEXT NOT NULL,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT,
			google_id TEXT,
			github_id TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image_path TEXT,
			theme TEXT,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS post_like (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS post_dislike (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS comment_like (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comments_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (comments_id) REFERENCES comments(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS comment_dislike (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comments_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (comments_id) REFERENCES comments(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("Erreur lors de la création des tables: %v", err)
		}
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

	hpassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = DB.Exec(`
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
	var storedPassword string
	err := DB.QueryRow(`SELECT password FROM users WHERE username = ?;`, username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return CheckPasswordHash(password, storedPassword), nil
}

func addPost(title, content, imagePath, theme string, userID int) error {
	_, err := DB.Exec(`
		INSERT INTO posts (title, content, image_path, theme, user_id) 
		VALUES (?, ?, ?, ?, ?);
	`, title, content, imagePath, theme, userID)
	return err
}

func getPosts(currentUserID int) ([]Post, error) {
	rows, err := DB.Query(`
		SELECT 
			posts.id,
			posts.title,
			posts.content,
			posts.image_path,
			posts.theme,
			COALESCE(users.full_name, users.username, 'Utilisateur'),
			COUNT(DISTINCT post_like.id) as likes,
			COUNT(DISTINCT post_dislike.id) as dislikes,
			EXISTS(SELECT 1 FROM post_like ul WHERE ul.post_id = posts.id AND ul.user_id = ?) as user_liked,
			EXISTS(SELECT 1 FROM post_dislike ud WHERE ud.post_id = posts.id AND ud.user_id = ?) as user_disliked
		FROM posts
		LEFT JOIN post_like ON posts.id = post_like.post_id
		LEFT JOIN post_dislike ON posts.id = post_dislike.post_id
		LEFT JOIN users ON posts.user_id = users.id
		GROUP BY posts.id;
	`, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var id, likes, dislikes int
		var userLiked, userDisliked bool
		var title, content, theme, username string
		var imagePath sql.NullString

		if err := rows.Scan(&id, &title, &content, &imagePath, &theme, &username, &likes, &dislikes, &userLiked, &userDisliked); err != nil {
			return nil, err
		}

		post := Post{
			ID:           id,
			Title:        title,
			Content:      content,
			Theme:        theme,
			Username:     username,
			Likes:        likes,
			Dislikes:     dislikes,
			UserLiked:    userLiked,
			UserDisliked: userDisliked,
		}
		if imagePath.Valid {
			post.ImagePath = imagePath.String
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func addComment(postID int, userID int, content string) error {
	_, err := DB.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?);
	`, postID, userID, content)
	return err
}

func getComments(postID int, currentUserID int) ([]Comment, error) {
	query := `
   SELECT 
      comments.id, 
      COALESCE(users.username, 'Anonyme') as username, 
      comments.content,
      COUNT(DISTINCT comment_like.id) as likes,
      COUNT(DISTINCT comment_dislike.id) as dislikes,
      MAX(CASE WHEN comment_like.user_id = ? THEN 1 ELSE 0 END) as user_liked,
      MAX(CASE WHEN comment_dislike.user_id = ? THEN 1 ELSE 0 END) as user_disliked
   FROM comments
   LEFT JOIN users ON comments.user_id = users.id
   LEFT JOIN comment_like ON comments.id = comment_like.comments_id
   LEFT JOIN comment_dislike ON comments.id = comment_dislike.comments_id
   WHERE comments.post_id = ?
   GROUP BY comments.id;
`

	rows, err := DB.Query(query, currentUserID, currentUserID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.Username, &c.Content, &c.Likes, &c.Dislikes, &c.UserLiked, &c.UserDisliked); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func deletePost(id int) error {
	_, err := DB.Exec(`DELETE FROM posts WHERE id = ?;`, id)
	return err
}

func userExistsByEmail(email string) (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

func insertGoogleUser(name, email, googleID string) error {
	_, err := DB.Exec(
		"INSERT INTO users (full_name, username, email, google_id) VALUES (?, ?, ?, ?)",
		name, name, email, googleID,
	)
	return err
}

func insertPostLike(postID int, userID int) error {
	_, err := DB.Exec("INSERT INTO post_like (post_id, user_id) VALUES (?, ?)", postID, userID)
	return err
}

func insertPostDislike(postID int, userID int) error {
	_, err := DB.Exec("INSERT INTO post_dislike (post_id, user_id) VALUES (?, ?)", postID, userID)
	return err
}

func insertCommentLike(commentID int, userID int) error {
	_, err := DB.Exec("INSERT INTO comment_like (comments_id, user_id) VALUES (?, ?)", commentID, userID)
	return err
}

func insertCommentDislike(commentID int, userID int) error {
	_, err := DB.Exec("INSERT INTO comment_dislike (comments_id, user_id) VALUES (?, ?)", commentID, userID)
	return err
}

func deleteCommentLike(commentID int, userID int) error {
	_, err := DB.Exec("DELETE FROM comment_like WHERE comments_id = ? AND user_id = ?;", commentID, userID)
	return err
}

func deleteCommentDislike(commentID int, userID int) error {
	_, err := DB.Exec("DELETE FROM comment_dislike WHERE comments_id = ? AND user_id = ?;", commentID, userID)
	return err
}

func deletePostLike(postID int, userID int) error {
	_, err := DB.Exec("DELETE FROM post_like WHERE post_id = ? AND user_id = ?;", postID, userID)
	return err
}

func deletePostDislike(postID int, userID int) error {
	_, err := DB.Exec("DELETE FROM post_dislike WHERE post_id = ? AND user_id = ?;", postID, userID)
	return err
}

func editComment(commentID int, content string, userID int) error {
	_, err := DB.Exec("UPDATE comments SET content = ? WHERE id = ? AND user_id = ?;", content, commentID, userID)
	return err
}

func editPostWithImage(id int, title, content, theme, imagePath string) error {
	_, err := DB.Exec("UPDATE posts SET title = ?, content = ?, theme = ?, image_path = ? WHERE id = ?;", title, content, theme, imagePath, id)
	return err
}

func editPostWithoutImage(id int, title, content, theme string) error {
	_, err := DB.Exec("UPDATE posts SET title = ?, content = ?, theme = ? WHERE id = ?;", title, content, theme, id)
	return err
}

// github auth
func updateGitHubID(email, githubID string) error {
	_, err := DB.Exec(
		"UPDATE users SET github_id = ? WHERE email = ? AND (github_id IS NULL OR github_id = '');",
		githubID, email,
	)
	return err
}

func insertGitHubUser(name, email, githubID string) error {
	_, err := DB.Exec(
		"INSERT INTO users (full_name, username, email, github_id) VALUES (?, ?, ?, ?)",
		name, name, email, githubID,
	)
	return err
}
