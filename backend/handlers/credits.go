package handlers

import (
	"finquest/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) GetCredits(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	items := make([]models.Credit, 0)
	if err := h.DB.SelectContext(c.Request.Context(), &items,
		`SELECT * FROM credits WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	); err != nil {
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
	var cr models.Credit
	err := h.DB.QueryRowxContext(c.Request.Context(),
		`INSERT INTO credits (user_id, type, bank_name, total_amount, remaining_balance, interest_rate, monthly_payment, note)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *`,
		userID, req.Type, req.BankName, req.TotalAmount, req.RemainingBalance,
		req.InterestRate, req.MonthlyPayment, req.Note,
	).StructScan(&cr)
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
	var cr models.Credit
	err := h.DB.QueryRowxContext(c.Request.Context(), `
		UPDATE credits SET
			bank_name         = COALESCE($1, bank_name),
			total_amount      = COALESCE($2, total_amount),
			remaining_balance = COALESCE($3, remaining_balance),
			interest_rate     = COALESCE($4, interest_rate),
			monthly_payment   = COALESCE($5, monthly_payment),
			note              = COALESCE($6, note)
		WHERE id = $7 AND user_id = $8 RETURNING *`,
		req.BankName, req.TotalAmount, req.RemainingBalance,
		req.InterestRate, req.MonthlyPayment, req.Note, id, userID,
	).StructScan(&cr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "credit not found"})
		return
	}
	c.JSON(http.StatusOK, cr)
}

func (h *Handler) DeleteCredit(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	id := c.Param("id")
	res, err := h.DB.ExecContext(c.Request.Context(),
		`DELETE FROM credits WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "credit not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
