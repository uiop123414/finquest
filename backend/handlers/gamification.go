package handlers

import (
	"finquest/models"
	"finquest/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetGamificationProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var user models.User
	if err := h.DB.QueryRowxContext(c.Request.Context(),
		`SELECT * FROM users WHERE id = $1`, userID,
	).StructScan(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// All achievements with earned_at if unlocked
	achievements := make([]models.Achievement, 0)
	_ = h.DB.SelectContext(c.Request.Context(), &achievements, `
		SELECT a.id, a.code, a.name, a.description, ua.earned_at
		FROM achievements a
		LEFT JOIN user_achievements ua ON ua.achievement_id = a.id AND ua.user_id = $1
		ORDER BY ua.earned_at NULLS LAST, a.name`,
		userID,
	)

	c.JSON(http.StatusOK, gin.H{
		"xp_total":          user.XPTotal,
		"level":             user.Level,
		"level_progress_pct": services.LevelProgressPct(user.XPTotal),
		"achievements":      achievements,
	})
}
