//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"finquest/config"
	"finquest/db"
	"finquest/handlers"
	"finquest/middleware"

	"github.com/gin-gonic/gin"
)

// setupRouter connects to TEST_DATABASE_URL and returns a configured router + cleanup fn.
// Skips the test if TEST_DATABASE_URL is not set.
func setupRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	database := db.Connect(dsn)
	cfg := &config.Config{
		JWTSecret:    "test-secret",
		Port:         "8000",
		AnthropicKey: "",
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.New(database, cfg)

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	p := api.Group("/", middleware.AuthRequired(cfg.JWTSecret))
	p.GET("/transactions", h.GetTransactions)
	p.POST("/transactions", h.CreateTransaction)
	p.GET("/analytics/summary", h.GetSummary)
	p.GET("/gamification/profile", h.GetGamificationProfile)

	cleanup := func() {
		database.MustExec("DELETE FROM users WHERE email LIKE 'testuser-%@example.com'")
		database.Close()
	}

	return r, cleanup
}

// doJSON fires a JSON request against the router and returns the recorder.
func doJSON(r *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// decode unmarshals the recorder body into v; fails the test on error.
func decode(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, w.Body.String())
	}
}

// uniqueEmail returns a unique test email that the cleanup query will delete.
func uniqueEmail() string {
	return fmt.Sprintf("testuser-%d@example.com", time.Now().UnixNano())
}

// registerAndGetToken registers a new user and returns the access token.
func registerAndGetToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	w := doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{
		"email": uniqueEmail(), "password": "password123",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("setup register: got %d: %s", w.Code, w.Body)
	}
	var resp map[string]interface{}
	decode(t, w, &resp)
	return resp["access_token"].(string)
}

// --- Auth ---

func TestRegister(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()

	email := uniqueEmail()

	// Success
	w := doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{
		"email": email, "password": "password123",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body)
	}
	var resp map[string]interface{}
	decode(t, w, &resp)
	if resp["access_token"] == nil {
		t.Fatal("register: missing access_token in response")
	}

	// Duplicate email → 409
	w = doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{
		"email": email, "password": "password123",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()

	email := uniqueEmail()
	doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{
		"email": email, "password": "password123",
	})

	// Success
	w := doJSON(r, "POST", "/api/v1/auth/login", "", map[string]string{
		"email": email, "password": "password123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", w.Code, w.Body)
	}
	var resp map[string]interface{}
	decode(t, w, &resp)
	if resp["access_token"] == nil {
		t.Fatal("login: missing access_token")
	}

	// Wrong password → 401
	w = doJSON(r, "POST", "/api/v1/auth/login", "", map[string]string{
		"email": email, "password": "wrongpassword",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: expected 401, got %d", w.Code)
	}
}

// --- Transactions ---

func TestCreateTransaction(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name:     "expense without category",
			body:     map[string]interface{}{"amount": 100.50, "type": "expense", "date": "2024-01-15", "note": "test-lunch"},
			wantCode: http.StatusCreated,
		},
		{
			name:     "income without category",
			body:     map[string]interface{}{"amount": 5000.0, "type": "income", "date": "2024-01-01"},
			wantCode: http.StatusCreated,
		},
		{
			name:     "invalid type",
			body:     map[string]interface{}{"amount": 50.0, "type": "transfer", "date": "2024-01-15"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid date format",
			body:     map[string]interface{}{"amount": 50.0, "type": "income", "date": "15-01-2024"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "zero amount",
			body:     map[string]interface{}{"amount": 0, "type": "income", "date": "2024-01-15"},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := doJSON(r, "POST", "/api/v1/transactions", token, tc.body)
			if w.Code != tc.wantCode {
				t.Errorf("expected %d, got %d: %s", tc.wantCode, w.Code, w.Body)
			}
		})
	}

	// No auth → 401
	w := doJSON(r, "POST", "/api/v1/transactions", "", map[string]interface{}{
		"amount": 50.0, "type": "income", "date": "2024-01-15",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no auth: expected 401, got %d", w.Code)
	}
}

func TestGetTransactions(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	// Seed one transaction
	doJSON(r, "POST", "/api/v1/transactions", token, map[string]interface{}{
		"amount": 200.0, "type": "income", "date": "2024-02-01", "note": "test-salary",
	})

	w := doJSON(r, "GET", "/api/v1/transactions", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get transactions: expected 200, got %d: %s", w.Code, w.Body)
	}

	var txs []interface{}
	decode(t, w, &txs)
	if len(txs) == 0 {
		t.Fatal("get transactions: expected at least 1 result, got empty array")
	}
}

// --- Analytics ---

func TestAnalyticsSummary(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	w := doJSON(r, "GET", "/api/v1/analytics/summary", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("analytics summary: expected 200, got %d: %s", w.Code, w.Body)
	}

	var summary map[string]interface{}
	decode(t, w, &summary)
	for _, key := range []string{"income", "expense", "balance", "by_category"} {
		if _, ok := summary[key]; !ok {
			t.Errorf("analytics summary: missing field %q", key)
		}
	}
}

// --- Gamification ---

func TestGamificationProfile(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	w := doJSON(r, "GET", "/api/v1/gamification/profile", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("gamification profile: expected 200, got %d: %s", w.Code, w.Body)
	}

	var profile map[string]interface{}
	decode(t, w, &profile)
	for _, key := range []string{"xp_total", "level", "achievements"} {
		if _, ok := profile[key]; !ok {
			t.Errorf("gamification profile: missing field %q", key)
		}
	}
}
