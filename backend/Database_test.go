package backend

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) func() {
	tmpDir := t.TempDir()
	dbPath = tmpDir + "/test.db"

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        full_name TEXT NOT NULL,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );`)
	db.Close()

	return func() {}
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

	err := insertUser("John Doe", "johndoe", "john@test.com", "password123", "password123")
	if err != nil {
		t.Errorf("insertUser échoué : %v", err)
	}
}

func TestInsertUser_PasswordMismatch(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	err := insertUser("John Doe", "johndoe2", "john2@test.com", "password123", "wrongpassword")
	if err == nil {
		t.Error("aurait dû retourner une erreur pour mots de passe différents")
	}
}

func TestInsertUser_DuplicateEmail(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "johndoe3", "same@test.com", "pass", "pass")
	err := insertUser("Jane Doe", "janedoe", "same@test.com", "pass", "pass")
	if err == nil {
		t.Error("aurait dû retourner une erreur pour email dupliqué")
	}
}

func TestInsertUser_DuplicateUsername(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "sameuser", "john3@test.com", "pass", "pass")
	err := insertUser("Jane Doe", "sameuser", "jane3@test.com", "pass", "pass")
	if err == nil {
		t.Error("aurait dû retourner une erreur pour username dupliqué")
	}
}

func TestLoginUser_Success(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "loginuser", "login@test.com", "mypassword", "mypassword")

	ok, err := loginUser("loginuser", "mypassword")
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

	insertUser("John Doe", "loginuser2", "login2@test.com", "mypassword", "mypassword")

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
