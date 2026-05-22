package handlers

import (
	"finquest/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetDeposits(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	items := make([]models.Deposit, 0)
	if err := h.DB.SelectContext(c.Request.Context(), &items,
		`SELECT * FROM deposits WHERE user_id = $1 ORDER BY end_date`, userID,
	); err != nil {
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
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
		return
	}
	var dep models.Deposit
	err = h.DB.QueryRowxContext(c.Request.Context(),
		`INSERT INTO deposits (user_id, bank_name, amount, interest_rate, start_date, end_date, note)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *`,
		userID, req.BankName, req.Amount, req.InterestRate, startDate, endDate, req.Note,
	).StructScan(&dep)
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
	var dep models.Deposit
	err := h.DB.QueryRowxContext(c.Request.Context(), `
		UPDATE deposits SET
			bank_name     = COALESCE($1, bank_name),
			amount        = COALESCE($2, amount),
			interest_rate = COALESCE($3, interest_rate),
			start_date    = COALESCE($4::date, start_date),
			end_date      = COALESCE($5::date, end_date),
			note          = COALESCE($6, note)
		WHERE id = $7 AND user_id = $8 RETURNING *`,
		req.BankName, req.Amount, req.InterestRate, req.StartDate, req.EndDate, req.Note, id, userID,
	).StructScan(&dep)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
		return
	}
	c.JSON(http.StatusOK, dep)
}

func (h *Handler) DeleteDeposit(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	id := c.Param("id")
	res, err := h.DB.ExecContext(c.Request.Context(),
		`DELETE FROM deposits WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "deposit not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
