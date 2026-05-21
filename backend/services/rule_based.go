package services

import "strings"

// keyword → category name
var rules = map[string]string{
	"магнит":      "Еда",
	"пятёрочка":   "Еда",
	"пятерочка":   "Еда",
	"перекрёсток": "Еда",
	"перекресток": "Еда",
	"вкусвилл":    "Еда",
	"вкусно":      "Еда",
	"ресторан":    "Еда",
	"кафе":        "Еда",
	"кофе":        "Еда",
	"metro":       "Транспорт",
	"метро":       "Транспорт",
	"такси":       "Транспорт",
	"яндекс такси": "Транспорт",
	"убер":        "Транспорт",
	"uber":        "Транспорт",
	"автобус":     "Транспорт",
	"жкх":         "ЖКХ",
	"коммунальные": "ЖКХ",
	"электричество": "ЖКХ",
	"аптека":      "Здоровье",
	"больница":    "Здоровье",
	"клиника":     "Здоровье",
	"кино":        "Развлечения",
	"театр":       "Развлечения",
	"игры":        "Развлечения",
	"зарплата":    "Зарплата",
	"оклад":       "Зарплата",
}

// RuleBasedCategorize returns category name for a description or empty string.
func RuleBasedCategorize(description string) string {
	lower := strings.ToLower(description)
	for keyword, category := range rules {
		if strings.Contains(lower, keyword) {
			return category
		}
	}
	return ""
}
