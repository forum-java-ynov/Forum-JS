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

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("attendu 405, reçu %d", rr.Code)
	}
}

func TestLogin_ChampsVides(t *testing.T) {
	body := strings.NewReader("username=&password=")
	req := httptest.NewRequest("POST", "/db/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("attendu 400, reçu %d", rr.Code)
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

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("attendu 401, reçu %d", rr.Code)
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

	if rr.Code != http.StatusBadRequest {
		t.Errorf("attendu 400, reçu %d", rr.Code)
	}
}

func TestRegister_PasswordMismatch(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	body := strings.NewReader("full_name=John&username=johndoe&email=john@test.com&password=abc&confirm_password=xyz")
	req := httptest.NewRequest("POST", "/db/register", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("attendu 400, reçu %d", rr.Code)
	}
}

func TestRegister_JSON(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	body := strings.NewReader(`{"full_name":"Jane","username":"janedoe","email":"jane@test.com","password":"pass","confirm_password":"pass"}`)
	req := httptest.NewRequest("POST", "/db/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	register(rr, req)

	if rr.Code != http.StatusSeeOther && rr.Code != http.StatusBadRequest {
		t.Errorf("code inattendu : %d", rr.Code)
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
