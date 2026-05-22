package handlers

import (
	"finquest/config"
	"finquest/repository"

	"github.com/jmoiron/sqlx"
)

// Handler хранит репозитории для большинства операций (Clean Architecture).
// DB оставлен для аналитических агрегатных запросов (analytics.go, ai.go)
// которые используют сложный SQL с JOIN/GROUP BY — их нецелесообразно строить через squirrel.
type Handler struct {
	DB           *sqlx.DB
	Transactions repository.TransactionRepo
	Goals        repository.GoalRepo
	Deposits     repository.DepositRepo
	Credits      repository.CreditRepo
	Categories   repository.CategoryRepo
	Users        repository.UserRepo
	Achievements repository.AchievementRepo
	XPEvents     repository.XPEventRepo
	Cfg          *config.Config
}

func New(
	db *sqlx.DB,
	txRepo repository.TransactionRepo,
	goalRepo repository.GoalRepo,
	depositRepo repository.DepositRepo,
	creditRepo repository.CreditRepo,
	catRepo repository.CategoryRepo,
	userRepo repository.UserRepo,
	achRepo repository.AchievementRepo,
	xpRepo repository.XPEventRepo,
	cfg *config.Config,
) *Handler {
	return &Handler{
		DB:           db,
		Transactions: txRepo,
		Goals:        goalRepo,
		Deposits:     depositRepo,
		Credits:      creditRepo,
		Categories:   catRepo,
		Users:        userRepo,
		Achievements: achRepo,
		XPEvents:     xpRepo,
		Cfg:          cfg,
	}
}
