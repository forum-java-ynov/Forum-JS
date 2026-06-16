package backend

import (
	"database/sql"
	"os"
	"testing"

	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) func() {
	os.Setenv("SESSION_KEY", "test-secret-key-for-testing-only")
	store = sessions.NewCookieStore([]byte("test-secret-key-for-testing-only"))
	tmpDir := t.TempDir()
	dbPath = tmpDir + "/test.db"

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	DB.Exec(`CREATE TABLE IF NOT EXISTS user_sessions (
		user_id INTEGER NOT NULL,
		session_token TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id)
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT,
		google_id TEXT,
		github_id TEXT,
		role TEXT NOT NULL DEFAULT 'user'
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		image_path TEXT,
		theme TEXT,
		user_id INTEGER NOT NULL
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS post_like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS post_dislike (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS comment_like (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		comments_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL
	);`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS comment_dislike (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		comments_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL
	);`)

	return func() {
		DB.Close()
	}
}

func TestCreateDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath = tmpDir + "/test.db"

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("impossible d'ouvrir la DB : %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("DB inaccessible : %v", err)
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("monpassword")
	if err != nil {
		t.Fatalf("HashPassword erreur : %v", err)
	}
	if hash == "monpassword" {
		t.Error("le mot de passe ne doit pas être stocké en clair")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	hash, _ := HashPassword("monpassword")

	if !CheckPasswordHash("monpassword", hash) {
		t.Error("aurait dû retourner true pour le bon mot de passe")
	}
	if CheckPasswordHash("mauvaispassword", hash) {
		t.Error("aurait dû retourner false pour un mauvais mot de passe")
	}
}

func TestInsertUser(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := insertUser("John Doe", "johndoe", "john@test.com", "Password123!", "Password123!")
	if err != nil {
		t.Errorf("insertUser échoué : %v", err)
	}
}

func TestInsertUser_PasswordMismatch(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := insertUser("John Doe", "johndoe2", "john2@test.com", "Password123!", "wrongpassword")
	if err == nil {
		t.Error("aurait dû retourner une erreur pour mots de passe différents")
	}
}

func TestInsertUser_DuplicateEmail(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "johndoe3", "same@test.com", "Password123!", "Password123!")
	err := insertUser("Jane Doe", "janedoe", "same@test.com", "Password123!", "Password123!")
	if err == nil {
		t.Error("aurait dû retourner une erreur pour email dupliqué")
	}
}

func TestInsertUser_DuplicateUsername(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "sameuser", "john3@test.com", "Password123!", "Password123!")
	err := insertUser("Jane Doe", "sameuser", "jane3@test.com", "Password123!", "Password123!")
	if err == nil {
		t.Error("aurait dû retourner une erreur pour username dupliqué")
	}
}

func TestInsertUser_AdminRole(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := insertUser("Admin", "admin", "admin@test.com", "Adminpass1!", "Adminpass1!")
	if err != nil {
		t.Fatalf("insertUser admin échoué : %v", err)
	}

	isAdmin, err := getUserRole(1)
	if err != nil {
		t.Fatalf("getUserRole erreur : %v", err)
	}
	if !isAdmin {
		t.Error("l'utilisateur admin devrait avoir le rôle admin")
	}
}

func TestLoginUser_Success(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "loginuser", "login@test.com", "Password123!", "Password123!")

	ok, err := loginUser("loginuser", "Password123!")
	if err != nil {
		t.Fatalf("loginUser erreur : %v", err)
	}
	if !ok {
		t.Error("aurait dû retourner true pour des identifiants corrects")
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "loginuser2", "login2@test.com", "Password123!", "Password123!")

	ok, err := loginUser("loginuser2", "wrongpassword")
	if err != nil {
		t.Fatalf("loginUser erreur : %v", err)
	}
	if ok {
		t.Error("aurait dû retourner false pour un mauvais mot de passe")
	}
}

func TestLoginUser_UserNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ok, err := loginUser("inexistant", "password")
	if err != nil {
		t.Fatalf("loginUser erreur : %v", err)
	}
	if ok {
		t.Error("aurait dû retourner false pour un utilisateur inexistant")
	}
}

func TestAddPost(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "poster", "poster@test.com", "Password123!", "Password123!")

	err := addPost("Mon titre", "Mon contenu", "", "tech", 1)
	if err != nil {
		t.Errorf("addPost échoué : %v", err)
	}
}

func TestAddComment(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "commenter", "commenter@test.com", "Password123!", "Password123!")
	addPost("Titre", "Contenu", "", "tech", 1)

	err := addComment(1, 1, "Mon commentaire")
	if err != nil {
		t.Errorf("addComment échoué : %v", err)
	}
}

func TestGetUserRole_NotAdmin(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("User Normal", "normaluser", "normal@test.com", "Password123!", "Password123!")

	isAdmin, err := getUserRole(1)
	if err != nil {
		t.Fatalf("getUserRole erreur : %v", err)
	}
	if isAdmin {
		t.Error("un utilisateur normal ne devrait pas être admin")
	}
}
