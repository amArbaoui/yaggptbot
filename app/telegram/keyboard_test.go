package telegram

import (
	"amArbaoui/yaggptbot/app/config"
	"maps"
	"slices"
	"testing"
)

func TestModelKeyboardIsSortedAndMarksCurrent(t *testing.T) {
	models := slices.Sorted(maps.Keys(config.ModelMap))
	current := models[len(models)-1]
	keyboard := NewModelKeyboard(1, current)
	if len(keyboard.InlineKeyboard) != len(models) {
		t.Fatalf("expected %d rows, got %d", len(models), len(keyboard.InlineKeyboard))
	}
	for i, row := range keyboard.InlineKeyboard {
		expected := models[i]
		if models[i] == current {
			expected = "✅ " + current
		}
		if row[0].Text != expected {
			t.Errorf("row %d: expected %q, got %q", i, expected, row[0].Text)
		}
	}
}
