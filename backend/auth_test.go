package backend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogin_MethodGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/db/login", nil)
	rr := httptest.NewRecorder()
	login(rr, req)

	// login redirige maintenant au lieu de retourner 405
	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
}

func TestLogin_ChampsVides(t *testing.T) {
	body := strings.NewReader("username=&password=")
	req := httptest.NewRequest("POST", "/db/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	login(rr, req)

	// redirige vers /login?error=... au lieu de 400
	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "error") {
		t.Errorf("attendu redirection avec error, reçu : %s", location)
	}
}

func TestLogin_IdentifiantsInvalides(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	body := strings.NewReader("username=inexistant&password=mauvais")
	req := httptest.NewRequest("POST", "/db/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "error") {
		t.Errorf("attendu redirection avec error, reçu : %s", location)
	}
}

func TestLogin_Success(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	insertUser("John Doe", "testlogin", "testlogin@test.com", "Password123!", "Password123!")

	body := strings.NewReader("username=testlogin&password=Password123!")
	req := httptest.NewRequest("POST", "/db/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if strings.Contains(location, "error") {
		t.Errorf("ne devrait pas contenir error, reçu : %s", location)
	}
}

func TestRegister_MethodGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/db/register", nil)
	rr := httptest.NewRecorder()
	register(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("attendu 405, reçu %d", rr.Code)
	}
}

func TestRegister_ChampsVides(t *testing.T) {
	body := strings.NewReader("full_name=&username=&email=&password=&confirm_password=")
	req := httptest.NewRequest("POST", "/db/register", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	register(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "error") {
		t.Errorf("attendu redirection avec error, reçu : %s", location)
	}
}

func TestRegister_PasswordMismatch(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	body := strings.NewReader("full_name=John&username=johndoe&email=john@test.com&password=Password123!&confirm_password=Password456!")
	req := httptest.NewRequest("POST", "/db/register", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	register(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "error") {
		t.Errorf("attendu redirection avec error, reçu : %s", location)
	}
}

func TestRegister_Success(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	body := strings.NewReader("full_name=Jane Doe&username=janedoe&email=jane@test.com&password=Password123!&confirm_password=Password123!")
	req := httptest.NewRequest("POST", "/db/register", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	register(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("attendu 303, reçu %d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "success") {
		t.Errorf("attendu redirection avec success, reçu : %s", location)
	}
}

func TestDecodeRequest_Form(t *testing.T) {
	body := strings.NewReader("username=testuser&password=testpass")
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var data LoginData
	err := decodeRequest(req, &data)
	if err != nil {
		t.Fatalf("decodeRequest erreur : %v", err)
	}
	if data.Username != "testuser" {
		t.Errorf("attendu 'testuser', reçu '%s'", data.Username)
	}
	if data.Password != "testpass" {
		t.Errorf("attendu 'testpass', reçu '%s'", data.Password)
	}
}

func TestDecodeRequest_JSON(t *testing.T) {
	body := strings.NewReader(`{"username":"testuser","password":"testpass"}`)
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/json")

	var data LoginData
	err := decodeRequest(req, &data)
	if err != nil {
		t.Fatalf("decodeRequest erreur : %v", err)
	}
	if data.Username != "testuser" {
		t.Errorf("attendu 'testuser', reçu '%s'", data.Username)
	}
}
