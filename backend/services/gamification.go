package services

import (
	"context"
	"finquest/repository"

	"github.com/google/uuid"
)

func CalculateLevel(xp int) int {
	if xp < 100 {
		return 1
	}
	return xp/100 + 1
}

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

// AwardXP добавляет XP, обновляет уровень и проверяет ачивки.
func AwardXP(
	ctx context.Context,
	userRepo repository.UserRepo,
	xpRepo repository.XPEventRepo,
	achRepo repository.AchievementRepo,
	txRepo repository.TransactionRepo,
	userID uuid.UUID,
	delta int,
	reason string,
) error {
	if err := xpRepo.Insert(ctx, userID, delta, reason); err != nil {
		return err
	}
	user, err := userRepo.AddXP(ctx, userID, delta)
	if err != nil {
		return err
	}
	txCount, _ := txRepo.CountByUser(ctx, userID)
	for _, check := range achievementChecks {
		if check.Condition(user.XPTotal, txCount) {
			_ = achRepo.Unlock(ctx, userID, check.Code)
		}
	}
	return nil
}
