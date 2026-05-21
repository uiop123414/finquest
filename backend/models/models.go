package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Email          string    `db:"email" json:"email"`
	HashedPassword string    `db:"hashed_password" json:"-"`
	XPTotal        int       `db:"xp_total" json:"xp_total"`
	Level          int       `db:"level" json:"level"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type Category struct {
	ID       uuid.UUID  `db:"id" json:"id"`
	UserID   *uuid.UUID `db:"user_id" json:"user_id"`
	Name     string     `db:"name" json:"name"`
	IsSystem bool       `db:"is_system" json:"is_system"`
}

type Transaction struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	Amount       float64    `db:"amount" json:"amount"`
	Type         string     `db:"type" json:"type"`
	CategoryID   *uuid.UUID `db:"category_id" json:"category_id"`
	Date         time.Time  `db:"date" json:"date"`
	Note         string     `db:"note" json:"note"`
	ExternalID   *string    `db:"external_id" json:"external_id,omitempty"`
	AIConfidence *float64   `db:"ai_confidence" json:"ai_confidence,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type XPEvent struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Delta     int       `db:"delta" json:"delta"`
	Reason    string    `db:"reason" json:"reason"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Achievement struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Code        string     `db:"code" json:"code"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	EarnedAt    *time.Time `db:"earned_at" json:"earned_at,omitempty"`
}

type Goal struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	UserID        uuid.UUID  `db:"user_id" json:"user_id"`
	Name          string     `db:"name" json:"name"`
	TargetAmount  float64    `db:"target_amount" json:"target_amount"`
	CurrentAmount float64    `db:"current_amount" json:"current_amount"`
	Deadline      time.Time  `db:"deadline" json:"deadline"`
	CompletedAt   *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}
