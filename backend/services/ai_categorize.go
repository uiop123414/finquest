package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TransactionForAI struct {
	Note string
}

type CategoryResult struct {
	CategoryName string
	Confidence   float64
}

type aiRequest struct {
	Model     string      `json:"model"`
	MaxTokens int         `json:"max_tokens"`
	Messages  []aiMessage `json:"messages"`
}

type aiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// CategorizeBatch sends transactions to Claude API and returns category names.
// Falls back to rule-based if API key is empty or request fails.
func CategorizeBatch(ctx context.Context, apiKey string, txs []TransactionForAI, categories []string) []CategoryResult {
	results := make([]CategoryResult, len(txs))

	if apiKey == "" {
		for i, tx := range txs {
			name := RuleBasedCategorize(tx.Note)
			if name == "" {
				name = "Прочее"
			}
			results[i] = CategoryResult{CategoryName: name, Confidence: 0.5}
		}
		return results
	}

	// Build prompt
	catList, _ := json.Marshal(categories)
	prompt := fmt.Sprintf(`Категоризируй транзакции. Категории: %s.
Формат ответа — JSON массив объектов {"category": "...", "confidence": 0.0-1.0}.
Примеры:
- "Магнит" → {"category": "Еда", "confidence": 0.95}
- "Яндекс Такси" → {"category": "Транспорт", "confidence": 0.97}

Транзакции:
`, string(catList))
	for i, tx := range txs {
		prompt += fmt.Sprintf("%d. %s\n", i+1, tx.Note)
	}

	body, _ := json.Marshal(aiRequest{
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 512,
		Messages:  []aiMessage{{Role: "user", Content: prompt}},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST",
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fallbackResults(txs)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var aiResp aiResponse
	if err := json.Unmarshal(raw, &aiResp); err != nil || len(aiResp.Content) == 0 {
		return fallbackResults(txs)
	}

	var parsed []struct {
		Category   string  `json:"category"`
		Confidence float64 `json:"confidence"`
	}
	// Extract JSON array from the response text
	text := aiResp.Content[0].Text
	start := bytes.IndexByte([]byte(text), '[')
	end := bytes.LastIndexByte([]byte(text), ']')
	if start < 0 || end < 0 {
		return fallbackResults(txs)
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &parsed); err != nil {
		return fallbackResults(txs)
	}

	for i := range results {
		if i < len(parsed) {
			results[i] = CategoryResult{
				CategoryName: parsed[i].Category,
				Confidence:   parsed[i].Confidence,
			}
		} else {
			results[i] = CategoryResult{CategoryName: "Прочее", Confidence: 0.5}
		}
	}
	return results
}

func fallbackResults(txs []TransactionForAI) []CategoryResult {
	results := make([]CategoryResult, len(txs))
	for i, tx := range txs {
		name := RuleBasedCategorize(tx.Note)
		if name == "" {
			name = "Прочее"
		}
		results[i] = CategoryResult{CategoryName: name, Confidence: 0.5}
	}
	return results
}
