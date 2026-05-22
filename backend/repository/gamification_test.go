package repository_test

import (
	"context"
	"errors"
	"finquest/models"
	"finquest/repository"
	"finquest/services"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ─── Моки ────────────────────────────────────────────────────────────────────

var errNotFound = errors.New("not found")

type mockUserRepo struct {
	users map[string]models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]models.User)}
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (models.User, error) {
	u, ok := m.users[email]
	if !ok {
		return models.User{}, errNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Create(_ context.Context, email, hashedPassword string) (models.User, error) {
	u := models.User{
		ID: uuid.New(), Email: email, HashedPassword: hashedPassword,
		XPTotal: 0, Level: 1, CreatedAt: time.Now(),
	}
	m.users[email] = u
	return u, nil
}

func (m *mockUserRepo) AddXP(_ context.Context, userID uuid.UUID, delta int) (models.User, error) {
	for email, u := range m.users {
		if u.ID == userID {
			u.XPTotal += delta
			u.Level = u.XPTotal/100 + 1
			m.users[email] = u
			return u, nil
		}
	}
	return models.User{}, errNotFound
}

func (m *mockUserRepo) GetProfile(_ context.Context, userID uuid.UUID) (models.User, error) {
	for _, u := range m.users {
		if u.ID == userID {
			return u, nil
		}
	}
	return models.User{}, errNotFound
}

type mockXPEventRepo struct {
	events []models.XPEvent
}

func (m *mockXPEventRepo) Insert(_ context.Context, userID uuid.UUID, delta int, reason string) error {
	m.events = append(m.events, models.XPEvent{
		ID: uuid.New(), UserID: userID, Delta: delta, Reason: reason, CreatedAt: time.Now(),
	})
	return nil
}

type mockAchievementRepo struct {
	unlocked map[string][]string
}

func newMockAchievementRepo() *mockAchievementRepo {
	return &mockAchievementRepo{unlocked: make(map[string][]string)}
}

func (m *mockAchievementRepo) ListForUser(_ context.Context, _ uuid.UUID) ([]models.Achievement, error) {
	return []models.Achievement{}, nil
}

func (m *mockAchievementRepo) Unlock(_ context.Context, userID uuid.UUID, code string) error {
	key := userID.String()
	for _, c := range m.unlocked[key] {
		if c == code {
			return nil
		}
	}
	m.unlocked[key] = append(m.unlocked[key], code)
	return nil
}

func (m *mockAchievementRepo) has(userID uuid.UUID, code string) bool {
	for _, c := range m.unlocked[userID.String()] {
		if c == code {
			return true
		}
	}
	return false
}

type mockTxRepo struct {
	store []models.Transaction
}

func (m *mockTxRepo) List(_ context.Context, userID uuid.UUID, _ repository.TransactionFilter) ([]models.Transaction, error) {
	var res []models.Transaction
	for _, tx := range m.store {
		if tx.UserID == userID {
			res = append(res, tx)
		}
	}
	return res, nil
}

func (m *mockTxRepo) Create(_ context.Context, userID uuid.UUID, amount float64, txType string, _ *uuid.UUID, date, note string) (models.Transaction, error) {
	tx := models.Transaction{ID: uuid.New(), UserID: userID, Amount: amount, Type: txType, Note: note, CreatedAt: time.Now()}
	if date != "" {
		tx.Date, _ = time.Parse("2006-01-02", date)
	}
	m.store = append(m.store, tx)
	return tx, nil
}

func (m *mockTxRepo) Update(_ context.Context, _, _ string, _ *float64, _, _, _, _ *string) (models.Transaction, error) {
	return models.Transaction{}, errNotFound
}

func (m *mockTxRepo) Delete(_ context.Context, id, userID string) (int64, error) {
	for i, tx := range m.store {
		if tx.ID.String() == id && tx.UserID.String() == userID {
			m.store = append(m.store[:i], m.store[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

func (m *mockTxRepo) ImportOne(_ context.Context, userID uuid.UUID, amount float64, txType string, _ *uuid.UUID, _ interface{}, note, _ string, _ *float64) (bool, error) {
	m.store = append(m.store, models.Transaction{ID: uuid.New(), UserID: userID, Amount: amount, Type: txType, Note: note})
	return true, nil
}

func (m *mockTxRepo) CountByUser(_ context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, tx := range m.store {
		if tx.UserID == userID {
			count++
		}
	}
	return count, nil
}

// ─── Тесты AwardXP ───────────────────────────────────────────────────────────

func TestAwardXP_AddsXP(t *testing.T) {
	userRepo := newMockUserRepo()
	xpRepo := &mockXPEventRepo{}
	achRepo := newMockAchievementRepo()
	txRepo := &mockTxRepo{}
	ctx := context.Background()

	user, _ := userRepo.Create(ctx, "test@example.com", "hash")

	err := services.AwardXP(ctx, userRepo, xpRepo, achRepo, txRepo, user.ID, 50, "test")
	if err != nil {
		t.Fatalf("AwardXP error: %v", err)
	}

	updated, _ := userRepo.GetProfile(ctx, user.ID)
	if updated.XPTotal != 50 {
		t.Errorf("expected XPTotal=50, got %d", updated.XPTotal)
	}
}

func TestAwardXP_LevelUpAt100(t *testing.T) {
	userRepo := newMockUserRepo()
	ctx := context.Background()
	user, _ := userRepo.Create(ctx, "lvl@test.com", "hash")

	_ = services.AwardXP(ctx, userRepo, &mockXPEventRepo{}, newMockAchievementRepo(), &mockTxRepo{}, user.ID, 100, "grind")

	updated, _ := userRepo.GetProfile(ctx, user.ID)
	if updated.Level != 2 {
		t.Errorf("expected Level=2 at 100 XP, got %d", updated.Level)
	}
}

func TestAwardXP_RecordsEvent(t *testing.T) {
	userRepo := newMockUserRepo()
	xpRepo := &mockXPEventRepo{}
	ctx := context.Background()
	user, _ := userRepo.Create(ctx, "evt@test.com", "hash")

	_ = services.AwardXP(ctx, userRepo, xpRepo, newMockAchievementRepo(), &mockTxRepo{}, user.ID, 10, "transaction_added")

	if len(xpRepo.events) != 1 {
		t.Fatalf("expected 1 XP event, got %d", len(xpRepo.events))
	}
	if xpRepo.events[0].Reason != "transaction_added" {
		t.Errorf("wrong reason: %s", xpRepo.events[0].Reason)
	}
}

func TestAwardXP_UnlocksFirstTransaction(t *testing.T) {
	userRepo := newMockUserRepo()
	achRepo := newMockAchievementRepo()
	txRepo := &mockTxRepo{}
	ctx := context.Background()

	user, _ := userRepo.Create(ctx, "ach@test.com", "hash")
	txRepo.Create(ctx, user.ID, 100, "expense", nil, "2025-01-01", "coffee")

	_ = services.AwardXP(ctx, userRepo, &mockXPEventRepo{}, achRepo, txRepo, user.ID, 10, "transaction_added")

	if !achRepo.has(user.ID, "first_transaction") {
		t.Error("expected first_transaction achievement to be unlocked")
	}
}

func TestAwardXP_UnlocksLevel5(t *testing.T) {
	userRepo := newMockUserRepo()
	achRepo := newMockAchievementRepo()
	ctx := context.Background()
	user, _ := userRepo.Create(ctx, "lvl5@test.com", "hash")

	_ = services.AwardXP(ctx, userRepo, &mockXPEventRepo{}, achRepo, &mockTxRepo{}, user.ID, 400, "grind")

	if !achRepo.has(user.ID, "level_5") {
		t.Error("expected level_5 achievement to be unlocked at 400 XP")
	}
}

func TestMockTxRepo_ListFiltersUsers(t *testing.T) {
	repo := &mockTxRepo{}
	ctx := context.Background()

	u1, u2 := uuid.New(), uuid.New()
	repo.Create(ctx, u1, 100, "income", nil, "2025-01-01", "a")
	repo.Create(ctx, u2, 200, "expense", nil, "2025-01-02", "b")

	txs, _ := repo.List(ctx, u1, repository.TransactionFilter{})
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction for u1, got %d", len(txs))
	}
}

func TestMockTxRepo_CountAfterDelete(t *testing.T) {
	repo := &mockTxRepo{}
	ctx := context.Background()
	userID := uuid.New()

	tx, _ := repo.Create(ctx, userID, 100, "income", nil, "2025-01-01", "salary")
	n, _ := repo.Delete(ctx, tx.ID.String(), userID.String())
	if n != 1 {
		t.Errorf("expected 1 row affected, got %d", n)
	}
	count, _ := repo.CountByUser(ctx, userID)
	if count != 0 {
		t.Errorf("expected count=0 after delete, got %d", count)
	}
}
