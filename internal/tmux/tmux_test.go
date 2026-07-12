package tmux

import "testing"

func TestSessionName(t *testing.T) {
	tests := map[string]string{
		"novel":            "novel",
		"novel love story": "novel-love-story",
		"한글":               "devx",
		"..demo..":         "demo",
	}
	for input, want := range tests {
		if got := SessionName(input); got != want {
			t.Fatalf("SessionName(%q) = %q, want %q", input, got, want)
		}
	}
}
