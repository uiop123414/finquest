package services

import "testing"

func TestRuleBasedCategorize(t *testing.T) {
	cases := []struct {
		desc, want string
	}{
		{"Покупка в Магнит", "Еда"},
		{"Яндекс Такси поездка", "Транспорт"},
		{"Оплата ЖКХ за январь", "ЖКХ"},
		{"Аптека 36.6", "Здоровье"},
		{"Кино Синема", "Развлечения"},
		{"Перевод другу", ""},
	}
	for _, c := range cases {
		got := RuleBasedCategorize(c.desc)
		if got != c.want {
			t.Errorf("RuleBasedCategorize(%q) = %q, want %q", c.desc, got, c.want)
		}
	}
}
