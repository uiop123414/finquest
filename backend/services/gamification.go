package services

import (
	"context"
	"finquest/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CalculateLevel returns level for a given XP total.
func CalculateLevel(xp int) int {
	if xp < 100 {
		return 1
	}
	return xp/100 + 1
}

// LevelProgressPct returns progress % toward the next level.
func LevelProgressPct(xp int) float64 {
	return float64(xp%100) / 100.0 * 100
}

type AchievementCheck struct {
	Code      string
	Condition func(xp, txCount int) bool
}

var achievementChecks = []AchievementCheck{
	{"first_transaction", func(_, txCount int) bool { return txCount >= 1 }},
	{"ten_transactions", func(_, txCount int) bool { return txCount >= 10 }},
	{"hundred_transactions", func(_, txCount int) bool { return txCount >= 100 }},
	{"level_5", func(xp, _ int) bool { return CalculateLevel(xp) >= 5 }},
}

// AwardXP adds XP to user, updates level, and checks for new achievements.
func AwardXP(ctx context.Context, db *sqlx.DB, userID uuid.UUID, delta int, reason string) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert xp event
	_, err = tx.ExecContext(ctx,
		`INSERT INTO xp_events (user_id, delta, reason) VALUES ($1, $2, $3)`,
		userID, delta, reason,
	)
	if err != nil {
		return err
	}

	// Update user xp + level
	var user models.User
	err = tx.GetContext(ctx, &user,
		`UPDATE users SET xp_total = xp_total + $1, level = (xp_total + $1) / 100 + 1
         WHERE id = $2 RETURNING *`,
		delta, userID,
	)
	if err != nil {
		return err
	}

	// Count transactions for achievement checks
	var txCount int
	_ = tx.GetContext(ctx, &txCount,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1`, userID,
	)

	// Check and unlock achievements
	for _, check := range achievementChecks {
		if !check.Condition(user.XPTotal, txCount) {
			continue
		}
		// Try to insert; ignore if already earned
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO user_achievements (user_id, achievement_id)
			SELECT $1, id FROM achievements WHERE code = $2
			ON CONFLICT DO NOTHING`,
			userID, check.Code,
		)
	}

	return tx.Commit()
}
