package handlers

import (
	"finquest/models"
	"finquest/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetGoals(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	goals := make([]models.Goal, 0)
	if err := h.DB.SelectContext(c.Request.Context(), &goals,
		`SELECT * FROM goals WHERE user_id = $1 ORDER BY completed_at NULLS FIRST, deadline`, userID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goals)
}

type createGoalRequest struct {
	Name          string  `json:"name" binding:"required"`
	TargetAmount  float64 `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64 `json:"current_amount"`
	Deadline      string  `json:"deadline" binding:"required"`
}

func (h *Handler) CreateGoal(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deadline, err := time.Parse("2006-01-02", req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	var goal models.Goal
	err = h.DB.QueryRowxContext(c.Request.Context(),
		`INSERT INTO goals (user_id, name, target_amount, current_amount, deadline)
         VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		userID, req.Name, req.TargetAmount, req.CurrentAmount, deadline,
	).StructScan(&goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// first_goal achievement + XP
	_ = services.AwardXP(c.Request.Context(), h.DB, userID, 20, "goal_created")
	_, _ = h.DB.ExecContext(c.Request.Context(), `
		INSERT INTO user_achievements (user_id, achievement_id)
		SELECT $1, id FROM achievements WHERE code = 'first_goal'
		ON CONFLICT DO NOTHING`, userID,
	)

	c.JSON(http.StatusCreated, goal)
}

type updateGoalRequest struct {
	Name          *string  `json:"name"`
	TargetAmount  *float64 `json:"target_amount"`
	CurrentAmount *float64 `json:"current_amount"`
	Deadline      *string  `json:"deadline"`
	// completed: true = mark done, false = reopen
	Completed *bool `json:"completed"`
}

func (h *Handler) UpdateGoal(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	goalID := c.Param("id")

	var req updateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var deadlineStr *string
	if req.Deadline != nil {
		t, err := time.Parse("2006-01-02", *req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline format"})
			return
		}
		s := t.Format("2006-01-02")
		deadlineStr = &s
	}

	var goal models.Goal
	var err error

	if req.Completed != nil {
		var completedAt *time.Time
		if *req.Completed {
			now := time.Now()
			completedAt = &now
		}
		err = h.DB.QueryRowxContext(c.Request.Context(), `
			UPDATE goals SET
				name           = COALESCE($1, name),
				target_amount  = COALESCE($2, target_amount),
				current_amount = COALESCE($3, current_amount),
				deadline       = COALESCE($4::date, deadline),
				completed_at   = $5
			WHERE id = $6 AND user_id = $7
			RETURNING *`,
			req.Name, req.TargetAmount, req.CurrentAmount, deadlineStr, completedAt, goalID, userID,
		).StructScan(&goal)
	} else {
		err = h.DB.QueryRowxContext(c.Request.Context(), `
			UPDATE goals SET
				name           = COALESCE($1, name),
				target_amount  = COALESCE($2, target_amount),
				current_amount = COALESCE($3, current_amount),
				deadline       = COALESCE($4::date, deadline)
			WHERE id = $5 AND user_id = $6
			RETURNING *`,
			req.Name, req.TargetAmount, req.CurrentAmount, deadlineStr, goalID, userID,
		).StructScan(&goal)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *Handler) DeleteGoal(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	goalID := c.Param("id")

	res, err := h.DB.ExecContext(c.Request.Context(),
		`DELETE FROM goals WHERE id = $1 AND user_id = $2`, goalID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
