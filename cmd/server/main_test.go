// cmd/server/main_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"go-jwt-api/internal/handlers"
	"go-jwt-api/internal/middleware"
)

// setupRouter はテスト用にルーターを構築します
func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.Handle("/protected",
		middleware.JwtMiddleware(http.HandlerFunc(Protected)),
	).Methods("GET")
	return r
}

func TestLogin_Success(t *testing.T) {
	creds := map[string]string{"username": "user1", "password": "password123"}
	body, _ := json.Marshal(creds)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if token, ok := resp["token"]; !ok || token == "" {
		t.Fatal("token not present in response")
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	creds := map[string]string{"username": "user1", "password": "wrong"}
	body, _ := json.Marshal(creds)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestProtected_Authorized(t *testing.T) {
	// /login で有効なトークンを取得
	creds := map[string]string{"username": "user1", "password": "password123"}
	loginBody, _ := json.Marshal(creds)
	loginReq := httptest.NewRequest("POST", "/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRec := httptest.NewRecorder()
	setupRouter().ServeHTTP(loginRec, loginReq)

	var loginResp map[string]string
	_ = json.NewDecoder(loginRec.Body).Decode(&loginResp)
	token := loginResp["token"]

	// トークンをセットして /protected へアクセス
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	setupRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	expected := "Hello, user1! This is a protected endpoint."
	if body := rr.Body.String(); body != expected {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestProtected_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()
	setupRouter().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}
