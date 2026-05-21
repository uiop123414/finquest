package services

import "testing"

func TestCalculateLevel(t *testing.T) {
	cases := []struct{ xp, want int }{
		{0, 1},
		{99, 1},
		{100, 2},
		{199, 2},
		{400, 5},
		{500, 6},
	}
	for _, c := range cases {
		got := CalculateLevel(c.xp)
		if got != c.want {
			t.Errorf("CalculateLevel(%d) = %d, want %d", c.xp, got, c.want)
		}
	}
}

func TestLevelProgressPct(t *testing.T) {
	if got := LevelProgressPct(150); got != 50.0 {
		t.Errorf("expected 50.0, got %f", got)
	}
	if got := LevelProgressPct(0); got != 0.0 {
		t.Errorf("expected 0.0, got %f", got)
	}
}
