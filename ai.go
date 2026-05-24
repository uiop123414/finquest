package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAIAdvice returns an AI-generated financial analysis (non-streaming).
// Includes: last-30-day transactions, deposits, credits, goals.
func (h *Handler) GetAIAdvice(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	ctx := c.Request.Context()

	// ── 1. Transactions (last 30 days) ───────────────────────────────────────
	type catRow struct {
		Category string  `db:"category"`
		Amount   float64 `db:"amount"`
		TxType   string  `db:"type"`
	}
	var txRows []catRow
	_ = h.DB.SelectContext(ctx, &txRows, `
		SELECT COALESCE(cat.name, 'Прочее') AS category,
		       SUM(t.amount) AS amount, t.type
		FROM transactions t
		LEFT JOIN categories cat ON cat.id = t.category_id
		WHERE t.user_id = $1 AND t.date >= NOW() - INTERVAL '30 days'
		GROUP BY cat.name, t.type ORDER BY t.type, amount DESC`, userID)

	var income, expense float64
	var expLines []string
	for _, r := range txRows {
		if r.TxType == "income" {
			income += r.Amount
		} else {
			expense += r.Amount
			expLines = append(expLines, fmt.Sprintf("  - %s: %.0f руб.", r.Category, r.Amount))
		}
	}

	// ── 2. Deposits ──────────────────────────────────────────────────────────
	type depRow struct {
		BankName     string  `db:"bank_name"`
		Amount       float64 `db:"amount"`
		InterestRate float64 `db:"interest_rate"`
		EndDate      string  `db:"end_date"`
	}
	var deps []depRow
	_ = h.DB.SelectContext(ctx, &deps, `
		SELECT bank_name, amount, interest_rate,
		       to_char(end_date, 'YYYY-MM-DD') AS end_date
		FROM deposits WHERE user_id = $1
		ORDER BY end_date`, userID)

	var totalDeposits float64
	var depLines []string
	for _, d := range deps {
		totalDeposits += d.Amount
		yearlyIncome := d.Amount * d.InterestRate / 100
		depLines = append(depLines, fmt.Sprintf(
			"  - %s: %.0f руб. под %.1f%% годовых (≈%.0f руб./год), до %s",
			d.BankName, d.Amount, d.InterestRate, yearlyIncome, d.EndDate,
		))
	}

	// ── 3. Credits ───────────────────────────────────────────────────────────
	type creditRow struct {
		CreditType       string  `db:"type"`
		BankName         string  `db:"bank_name"`
		TotalAmount      float64 `db:"total_amount"`
		RemainingBalance float64 `db:"remaining_balance"`
		InterestRate     float64 `db:"interest_rate"`
		MonthlyPayment   float64 `db:"monthly_payment"`
	}
	var crds []creditRow
	_ = h.DB.SelectContext(ctx, &crds, `
		SELECT type, bank_name, total_amount, remaining_balance, interest_rate, monthly_payment
		FROM credits WHERE user_id = $1`, userID)

	var totalDebt, totalMonthlyPayment float64
	var crdLines []string
	for _, cr := range crds {
		totalDebt += cr.RemainingBalance
		totalMonthlyPayment += cr.MonthlyPayment
		typeName := "Потребительский кредит"
		if cr.CreditType == "card" {
			typeName = "Кредитная карта"
		}
		crdLines = append(crdLines, fmt.Sprintf(
			"  - %s (%s): долг %.0f руб. / лимит %.0f руб., ставка %.1f%%, платёж %.0f руб./мес",
			typeName, cr.BankName, cr.RemainingBalance, cr.TotalAmount,
			cr.InterestRate, cr.MonthlyPayment,
		))
	}

	// ── 4. Goals ─────────────────────────────────────────────────────────────
	type goalRow struct {
		Name          string  `db:"name"`
		CurrentAmount float64 `db:"current_amount"`
		TargetAmount  float64 `db:"target_amount"`
		Deadline      string  `db:"deadline"`
	}
	var goals []goalRow
	_ = h.DB.SelectContext(ctx, &goals, `
		SELECT name, current_amount, target_amount,
		       to_char(deadline, 'YYYY-MM-DD') AS deadline
		FROM goals WHERE user_id = $1 AND completed_at IS NULL
		ORDER BY deadline LIMIT 5`, userID)

	var goalLines []string
	for _, g := range goals {
		pct := 0.0
		if g.TargetAmount > 0 {
			pct = g.CurrentAmount / g.TargetAmount * 100
		}
		goalLines = append(goalLines, fmt.Sprintf(
			"  - %s: %.0f/%.0f руб. (%.0f%%), дедлайн %s",
			g.Name, g.CurrentAmount, g.TargetAmount, pct, g.Deadline,
		))
	}

	// ── Build context string ─────────────────────────────────────────────────
	var sb strings.Builder
	fmt.Fprintf(&sb, "За последние 30 дней:\n- Доходы: %.0f руб.\n- Расходы: %.0f руб.\n", income, expense)

	if income > 0 {
		savingRate := (income - expense) / income * 100
		fmt.Fprintf(&sb, "- Норма сбережений: %.0f%%\n", savingRate)
	}

	if len(expLines) > 0 {
		fmt.Fprintf(&sb, "\nРасходы по категориям:\n%s\n", strings.Join(expLines, "\n"))
	}

	if len(depLines) > 0 {
		fmt.Fprintf(&sb, "\nБанковские депозиты (итого %.0f руб.):\n%s\n",
			totalDeposits, strings.Join(depLines, "\n"))
	} else {
		fmt.Fprintf(&sb, "\nБанковских депозитов нет.\n")
	}

	if len(crdLines) > 0 {
		fmt.Fprintf(&sb, "\nКредиты и кредитные карты (долг %.0f руб., ежемес. нагрузка %.0f руб.):\n%s\n",
			totalDebt, totalMonthlyPayment, strings.Join(crdLines, "\n"))
	} else {
		fmt.Fprintf(&sb, "\nКредитов и карт с долгом нет.\n")
	}

	if len(goalLines) > 0 {
		fmt.Fprintf(&sb, "\nАктивные финансовые цели:\n%s\n", strings.Join(goalLines, "\n"))
	}

	context := sb.String()

	// ── Choose provider: Anthropic → Gemini → rule-based ─────────────────────
	topic := c.DefaultQuery("topic", "general")
	prompt := buildPromptForTopic(context, topic)

	var advice string
	var err error
	switch {
	case h.Cfg.AnthropicKey != "":
		advice, err = callAnthropic(ctx, h.Cfg.AnthropicKey, prompt)
	case h.Cfg.GeminiKey != "":
		advice, err = callGemini(ctx, h.Cfg.GeminiKey, prompt)
	default:
		advice = buildRuleBasedAdvice(income, expense, totalDebt, totalMonthlyPayment, totalDeposits)
	}

	if err != nil {
		advice = buildRuleBasedAdvice(income, expense, totalDebt, totalMonthlyPayment, totalDeposits)
	}

	c.JSON(http.StatusOK, gin.H{"advice": advice, "context": context})
}

func buildPromptForTopic(financialContext, topic string) string {
	base := financialContext + "\n"
	suffix := "Пиши конкретно, с цифрами из контекста. Без markdown, простым текстом. На русском языке.\n"

	general := `Ты — персональный финансовый советник. На основе данных выше дай 4-5 конкретных советов:
1. Как улучшить баланс доходов и расходов
2. Стоит ли открыть/закрыть депозиты или перераспределить их
3. Как оптимизировать кредитную нагрузку (если есть)
4. Приоритеты для достижения финансовых целей
5. Один неочевидный совет под конкретную ситуацию`

	savings := `Ты — эксперт по личным сбережениям. На основе данных дай 4-5 конкретных советов:
1. Оцени текущую норму сбережений и назови целевую
2. Какие статьи расходов можно сократить на 10-20% прямо сейчас
3. Правило 50/30/20 — насколько текущий бюджет ему соответствует
4. Конкретный план: сколько откладывать в месяц и куда
5. Психологический совет: как не срываться с режима экономии`

	investments := `Ты — инвестиционный советник. На основе данных дай 4-5 конкретных рекомендаций:
1. Оцени текущие депозиты: выгодны ли ставки, стоит ли реинвестировать
2. Есть ли свободный капитал для инвестиций после погашения долгов
3. Какой инструмент подходит для этого профиля: ОФЗ, фонды, вклады
4. Как диверсифицировать накопления с учётом кредитной нагрузки
5. Конкретный шаг, который можно сделать уже в этом месяце`

	debt := `Ты — специалист по управлению долгами. На основе данных дай 4-5 конкретных советов:
1. Оцени кредитную нагрузку как процент от дохода (норма — до 30%)
2. Какой долг погашать первым: метод лавины или снежного кома
3. Стоит ли рефинансировать кредиты и при каких условиях
4. Как ускорить погашение без ущерба для качества жизни
5. Красные флаги: есть ли риск попасть в долговую ловушку`

	goals := `Ты — коуч по финансовым целям. На основе данных дай 4-5 конкретных рекомендаций:
1. Реалистичны ли текущие цели с учётом доходов и расходов
2. Какую цель приоритизировать и почему
3. Сколько нужно откладывать ежемесячно для каждой цели
4. Как ускорить достижение самой важной цели
5. Одна цель, которую стоит пересмотреть или разбить на этапы`

	instructions := map[string]string{
		"general":     general,
		"savings":     savings,
		"investments": investments,
		"debt":        debt,
		"goals":       goals,
	}

	instruction, ok := instructions[topic]
	if !ok {
		instruction = general
	}

	return base + instruction + "\n\n" + suffix
}

func callAnthropic(ctx context.Context, apiKey, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 600,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	raw, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(raw, &result); err != nil || len(result.Content) == 0 {
		return "", fmt.Errorf("invalid anthropic response")
	}
	return result.Content[0].Text, nil
}

func callGemini(ctx context.Context, apiKey, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	})
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct{ Text string `json:"text"` } `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	raw, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(raw, &result); err != nil ||
		len(result.Candidates) == 0 ||
		len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("invalid gemini response")
	}
	return result.Candidates[0].Content.Parts[0].Text, nil
}

func buildRuleBasedAdvice(income, expense, totalDebt, monthlyPayment, totalDeposits float64) string {
	var parts []string

	savingRate := 0.0
	if income > 0 {
		savingRate = (income - expense) / income * 100
	}

	switch {
	case income == 0 && expense == 0:
		return "За последние 30 дней транзакций не найдено. Добавьте расходы и доходы, чтобы получить персональный анализ."
	case savingRate >= 30:
		parts = append(parts, fmt.Sprintf("Отличный результат! Вы сберегаете %.0f%% дохода. Направьте излишек на пополнение депозита или досрочное погашение кредита.", savingRate))
	case savingRate >= 10:
		parts = append(parts, fmt.Sprintf("Норма сбережений %.0f%% — хороший показатель. Постарайтесь увеличить её до 20%%, сократив необязательные расходы.", savingRate))
	case savingRate >= 0:
		parts = append(parts, "Расходы почти равны доходам. Проанализируйте статьи расходов и найдите 10–15% для сокращения.")
	default:
		parts = append(parts, fmt.Sprintf("Расходы превышают доходы на %.0f руб. Это требует немедленного внимания — составьте бюджет.", expense-income))
	}

	if totalDebt > 0 && monthlyPayment > 0 && income > 0 {
		debtLoad := monthlyPayment / income * 100
		if debtLoad > 40 {
			parts = append(parts, fmt.Sprintf("Кредитная нагрузка %.0f%% от дохода — это высокий уровень (норма до 30%%). Рассмотрите рефинансирование.", debtLoad))
		} else {
			parts = append(parts, fmt.Sprintf("Кредитная нагрузка %.0f%% от дохода — в пределах нормы.", debtLoad))
		}
	}

	if totalDeposits == 0 && savingRate > 10 {
		parts = append(parts, "У вас нет активных депозитов. Рассмотрите открытие вклада для защиты сбережений от инфляции.")
	}

	if len(parts) == 0 {
		parts = append(parts, "Для получения детального анализа добавьте ANTHROPIC_API_KEY в настройки.")
	}

	return strings.Join(parts, " ")
}

// AIChat — streaming chat endpoint (unchanged)
func (h *Handler) AIChat(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(string)

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
		"Ты финансовый советник. За последние 30 дней пользователь заработал %.2f руб. и потратил %.2f руб. Отвечай кратко и по делу на русском языке.",
		income, expense,
	)

	if h.Cfg.AnthropicKey == "" {
		c.JSON(http.StatusOK, gin.H{"reply": "AI недоступен: не задан ANTHROPIC_API_KEY"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	body, _ := json.Marshal(map[string]interface{}{
		"model":      "claude-haiku-4-5-20251001",
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
