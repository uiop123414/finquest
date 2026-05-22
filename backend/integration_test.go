//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"finquest/config"
	"finquest/db"
	"finquest/handlers"
	"finquest/middleware"
	"finquest/repository"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// testDSN — DSN контейнера, заполняется в TestMain один раз для всех тестов.
var testDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgc, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("finquest_test"),
		tcpostgres.WithUsername("finquest"),
		tcpostgres.WithPassword("finquest"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}
	defer func() {
		if err := pgc.Terminate(ctx); err != nil {
			log.Printf("terminate container: %v", err)
		}
	}()

	testDSN, err = pgc.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("connection string: %v", err)
	}

	conn := db.Connect(testDSN)
	if err := applyMigrations(conn); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	conn.Close()

	os.Exit(m.Run())
}

// applyMigrations выполняет структурные миграции (без seed-данных).
func applyMigrations(conn *sqlx.DB) error {
	files := []string{
		"db/migrations/001_init.up.sql",
		"db/migrations/003_goals_completed.up.sql",
		"db/migrations/004_investments_credits.up.sql",
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := conn.Exec(string(data)); err != nil {
			return fmt.Errorf("exec %s: %w", f, err)
		}
	}
	return nil
}

// ─── Вспомогательные функции ─────────────────────────────────────────────────

func setupRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()

	database := db.Connect(testDSN)
	cfg := &config.Config{JWTSecret: "test-secret", Port: "8000"}

	txRepo   := repository.NewTransactionRepo(database)
	goalRepo := repository.NewGoalRepo(database)
	depRepo  := repository.NewDepositRepo(database)
	crRepo   := repository.NewCreditRepo(database)
	catRepo  := repository.NewCategoryRepo(database)
	userRepo := repository.NewUserRepo(database)
	achRepo  := repository.NewAchievementRepo(database)
	xpRepo   := repository.NewXPEventRepo(database)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.New(database, txRepo, goalRepo, depRepo, crRepo, catRepo, userRepo, achRepo, xpRepo, cfg)

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

func decode(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, w.Body.String())
	}
}

func uniqueEmail() string {
	return fmt.Sprintf("testuser-%d@example.com", time.Now().UnixNano())
}

func registerAndGetToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	w := doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{
		"email": uniqueEmail(), "password": "password123",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: got %d: %s", w.Code, w.Body)
	}
	var resp map[string]interface{}
	decode(t, w, &resp)
	return resp["access_token"].(string)
}

// ─── Тесты ───────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	email := uniqueEmail()

	w := doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{"email": email, "password": "password123"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body)
	}
	var resp map[string]interface{}
	decode(t, w, &resp)
	if resp["access_token"] == nil {
		t.Fatal("missing access_token")
	}

	// Повторная регистрация — 409 Conflict
	w = doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{"email": email, "password": "password123"})
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate: expected 409, got %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	email := uniqueEmail()
	doJSON(r, "POST", "/api/v1/auth/register", "", map[string]string{"email": email, "password": "password123"})

	w := doJSON(r, "POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": "password123"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}

	// Неверный пароль — 401
	w = doJSON(r, "POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": "wrong"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreateTransaction(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{"expense ok",   map[string]interface{}{"amount": 100.50, "type": "expense",  "date": "2024-01-15"}, http.StatusCreated},
		{"income ok",    map[string]interface{}{"amount": 5000.0, "type": "income",   "date": "2024-01-01"}, http.StatusCreated},
		{"invalid type", map[string]interface{}{"amount": 50.0,   "type": "transfer", "date": "2024-01-15"}, http.StatusBadRequest},
		{"invalid date", map[string]interface{}{"amount": 50.0,   "type": "income",   "date": "15-01-2024"}, http.StatusBadRequest},
		{"zero amount",  map[string]interface{}{"amount": 0,      "type": "income",   "date": "2024-01-15"}, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := doJSON(r, "POST", "/api/v1/transactions", token, tc.body)
			if w.Code != tc.wantCode {
				t.Errorf("expected %d, got %d: %s", tc.wantCode, w.Code, w.Body)
			}
		})
	}

	// Без токена — 401
	w := doJSON(r, "POST", "/api/v1/transactions", "", map[string]interface{}{"amount": 50.0, "type": "income", "date": "2024-01-15"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no auth: expected 401, got %d", w.Code)
	}
}

func TestGetTransactions(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)
	doJSON(r, "POST", "/api/v1/transactions", token, map[string]interface{}{"amount": 200.0, "type": "income", "date": "2024-02-01"})

	w := doJSON(r, "GET", "/api/v1/transactions", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var txs []interface{}
	decode(t, w, &txs)
	if len(txs) == 0 {
		t.Fatal("expected at least 1 transaction")
	}
}

func TestAnalyticsSummary(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	w := doJSON(r, "GET", "/api/v1/analytics/summary", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var summary map[string]interface{}
	decode(t, w, &summary)
	for _, key := range []string{"income", "expense", "balance", "by_category"} {
		if _, ok := summary[key]; !ok {
			t.Errorf("missing field %q", key)
		}
	}
}

func TestGamificationProfile(t *testing.T) {
	r, cleanup := setupRouter(t)
	defer cleanup()
	token := registerAndGetToken(t, r)

	w := doJSON(r, "GET", "/api/v1/gamification/profile", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var profile map[string]interface{}
	decode(t, w, &profile)
	for _, key := range []string{"xp_total", "level", "achievements"} {
		if _, ok := profile[key]; !ok {
			t.Errorf("missing field %q", key)
		}
	}
}
