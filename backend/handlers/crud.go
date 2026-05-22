package handlers

import (
	"finquest/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ── Goals ─────────────────────────────────────────────────────────────────────

func (h *Handler) GetGoals(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	goals, err := h.Goals.List(c.Request.Context(), userID)
	if err != nil {
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
	if _, err := time.Parse("2006-01-02", req.Deadline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}
	goal, err := h.Goals.Create(c.Request.Context(), userID, req.Name, req.TargetAmount, req.CurrentAmount, req.Deadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = services.AwardXP(c.Request.Context(), h.Users, h.XPEvents, h.Achievements, h.Transactions, userID, 20, "goal_created")
	_ = h.Achievements.Unlock(c.Request.Context(), userID, "first_goal")
	c.JSON(http.StatusCreated, goal)
}

type updateGoalRequest struct {
	Name          *string  `json:"name"`
	TargetAmount  *float64 `json:"target_amount"`
	CurrentAmount *float64 `json:"current_amount"`
	Deadline      *string  `json:"deadline"`
	Completed     *bool    `json:"completed"`
}

func (h *Handler) UpdateGoal(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	goalID := c.Param("id")
	var req updateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Deadline != nil {
		if _, err := time.Parse("2006-01-02", *req.Deadline); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline format"})
			return
		}
	}
	var completedAt interface{}
	if req.Completed != nil {
		if *req.Completed {
			now := time.Now()
			completedAt = now
		} else {
			var t *time.Time
			completedAt = t
		}
	}
	goal, err := h.Goals.Update(c.Request.Context(), goalID, userID, req.Name, req.TargetAmount, req.CurrentAmount, req.Deadline, completedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	c.JSON(http.StatusOK, goal)
}

func (h *Handler) DeleteGoal(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	goalID := c.Param("id")
	n, err := h.Goals.Delete(c.Request.Context(), goalID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── Deposits ──────────────────────────────────────────────────────────────────

func (h *Handler) GetDeposits(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	items, err := h.Deposits.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

type createDepositRequest struct {
	BankName     string  `json:"bank_name" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	InterestRate float64 `json:"interest_rate" binding:"required,gte=0"`
	StartDate    string  `json:"start_date" binding:"required"`
	EndDate      string  `json:"end_date" binding:"required"`
	Note         string  `json:"note"`
}

func (h *Handler) CreateDeposit(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))
	var req createDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dep, err := h.Deposits.Create(c.Request.Context(), userID, req.BankName, req.Amount, req.InterestRate, req.StartDate, req.EndDate, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dep)
}

type updateDepositRequest struct {
	BankName     *string  `json:"bank_name"`
	Amount       *float64 `json:"amount"`
	InterestRate *float64 `json:"interest_rate"`
	StartDate    *string  `json:"start_date"`
	EndDate      *string  `json:"end_date"`
	Note         *string  `json:"note"`
}

func (h *Handler) UpdateDeposit(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	id := c.Param("id")
	var req updateDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dep, err := h.Deposits.Update(c.Request.Context(), id, userID, req.BankName, req.Amount, req.InterestRate, req.StartDate, req.EndDate, req.Note)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
		return
	}
	c.JSON(http.StatusOK, dep)
}

func (h *Handler) DeleteDeposit(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	id := c.Param("id")
	n, err := h.Deposits.Delete(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── Credits ───────────────────────────────────────────────────────────────────

func (h *Handler) GetCredits(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	items, err := h.Credits.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

type createCreditRequest struct {
	Type             string  `json:"type" binding:"required,oneof=consumer card"`
	BankName         string  `json:"bank_name" binding:"required"`
	TotalAmount      float64 `json:"total_amount" binding:"required,gt=0"`
	RemainingBalance float64 `json:"remaining_balance" binding:"gte=0"`
	InterestRate     float64 `json:"interest_rate" binding:"required,gte=0"`
	MonthlyPayment   float64 `json:"monthly_payment" binding:"gte=0"`
	Note             string  `json:"note"`
}

func (h *Handler) CreateCredit(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))
	var req createCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cr, err := h.Credits.Create(c.Request.Context(), userID, req.Type, req.BankName, req.TotalAmount, req.RemainingBalance, req.InterestRate, req.MonthlyPayment, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cr)
}

type updateCreditRequest struct {
	BankName         *string  `json:"bank_name"`
	TotalAmount      *float64 `json:"total_amount"`
	RemainingBalance *float64 `json:"remaining_balance"`
	InterestRate     *float64 `json:"interest_rate"`
	MonthlyPayment   *float64 `json:"monthly_payment"`
	Note             *string  `json:"note"`
}

func (h *Handler) UpdateCredit(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	id := c.Param("id")
	var req updateCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cr, err := h.Credits.Update(c.Request.Context(), id, userID, req.BankName, req.TotalAmount, req.RemainingBalance, req.InterestRate, req.MonthlyPayment, req.Note)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "credit not found"})
		return
	}
	c.JSON(http.StatusOK, cr)
}

func (h *Handler) DeleteCredit(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	id := c.Param("id")
	n, err := h.Credits.Delete(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "credit not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── Categories ────────────────────────────────────────────────────────────────

func (h *Handler) GetCategories(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	cats, err := h.Categories.ListForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

type createCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
	userID, _ := uuid.Parse(c.MustGet("userID").(string))
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := h.Categories.Create(c.Request.Context(), userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}
