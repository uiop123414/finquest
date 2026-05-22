package repository

import (
	"context"
	"finquest/models"

	"github.com/google/uuid"
)

// TransactionFilter — опциональные фильтры для списка транзакций.
type TransactionFilter struct {
	DateFrom   string
	DateTo     string
	CategoryID string
	Limit      int
	Offset     int
}

type TransactionRepo interface {
	List(ctx context.Context, userID uuid.UUID, f TransactionFilter) ([]models.Transaction, error)
	Create(ctx context.Context, userID uuid.UUID, amount float64, txType string, categoryID *uuid.UUID, date, note string) (models.Transaction, error)
	Update(ctx context.Context, id, userID string, amount *float64, txType, categoryID, date, note *string) (models.Transaction, error)
	Delete(ctx context.Context, id, userID string) (int64, error)
	ImportOne(ctx context.Context, userID uuid.UUID, amount float64, txType string, categoryID *uuid.UUID, date interface{}, note, externalID string, aiConfidence *float64) (bool, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
}

type GoalRepo interface {
	List(ctx context.Context, userID string) ([]models.Goal, error)
	Create(ctx context.Context, userID uuid.UUID, name string, target, current float64, deadline string) (models.Goal, error)
	Update(ctx context.Context, id, userID string, name *string, target, current *float64, deadline *string, completedAt interface{}) (models.Goal, error)
	Delete(ctx context.Context, id, userID string) (int64, error)
}

type DepositRepo interface {
	List(ctx context.Context, userID string) ([]models.Deposit, error)
	Create(ctx context.Context, userID uuid.UUID, bankName string, amount, rate float64, startDate, endDate, note string) (models.Deposit, error)
	Update(ctx context.Context, id, userID string, bankName *string, amount, rate *float64, startDate, endDate, note *string) (models.Deposit, error)
	Delete(ctx context.Context, id, userID string) (int64, error)
}

type CreditRepo interface {
	List(ctx context.Context, userID string) ([]models.Credit, error)
	Create(ctx context.Context, userID uuid.UUID, creditType, bankName string, total, remaining, rate, monthly float64, note string) (models.Credit, error)
	Update(ctx context.Context, id, userID string, bankName *string, total, remaining, rate, monthly *float64, note *string) (models.Credit, error)
	Delete(ctx context.Context, id, userID string) (int64, error)
}

type CategoryRepo interface {
	ListForUser(ctx context.Context, userID string) ([]models.Category, error)
	ListNamesForUser(ctx context.Context, userID uuid.UUID) ([]string, error)
	FindIDByName(ctx context.Context, name string, userID uuid.UUID) (uuid.UUID, error)
	Create(ctx context.Context, userID uuid.UUID, name string) (models.Category, error)
}

type UserRepo interface {
	FindByEmail(ctx context.Context, email string) (models.User, error)
	Create(ctx context.Context, email, hashedPassword string) (models.User, error)
	AddXP(ctx context.Context, userID uuid.UUID, delta int) (models.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (models.User, error)
}

type AchievementRepo interface {
	ListForUser(ctx context.Context, userID uuid.UUID) ([]models.Achievement, error)
	Unlock(ctx context.Context, userID uuid.UUID, code string) error
}

type XPEventRepo interface {
	Insert(ctx context.Context, userID uuid.UUID, delta int, reason string) error
}
