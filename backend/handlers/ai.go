package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type aiChatRequest struct {
	Message string `json:"message" binding:"required"`
}

func (h *Handler) AIChat(c *gin.Context) {
	var req aiChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(string)

	// Get user's financial context
	var income, expense float64
	_ = h.DB.QueryRowContext(c.Request.Context(), `
		SELECT
			COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1 AND date >= NOW() - INTERVAL '30 days'`,
		userID,
	).Scan(&income, &expense)

	systemPrompt := fmt.Sprintf(
		"Ты финансовый советник. За последние 30 дней пользователь заработал %.2f руб. и потратил %.2f руб. "+
			"Отвечай кратко и по делу на русском языке.",
		income, expense,
	)

	if h.Cfg.AnthropicKey == "" {
		c.JSON(http.StatusOK, gin.H{"reply": "AI недоступен: не задан ANTHROPIC_API_KEY"})
		return
	}

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	body, _ := json.Marshal(map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 512,
		"stream":     true,
		"system":     systemPrompt,
		"messages":   []map[string]string{{"role": "user", "content": req.Message}},
	})

	apiReq, _ := http.NewRequestWithContext(c.Request.Context(), "POST",
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	apiReq.Header.Set("x-api-key", h.Cfg.AnthropicKey)
	apiReq.Header.Set("anthropic-version", "2023-06-01")
	apiReq.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\": \"api error\"}\n\n")
		return
	}
	defer resp.Body.Close()

	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 256)
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Fprintf(w, "%s", buf[:n])
		}
		return err == nil
	})
}
