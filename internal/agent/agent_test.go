package agent

import "testing"

func TestGet(t *testing.T) {
	a, err := Get("claude")
	if err != nil {
		t.Fatal(err)
	}
	if a.Command != "claude" {
		t.Fatalf("command = %q", a.Command)
	}

	g, err := Get("gemini")
	if err != nil {
		t.Fatal(err)
	}
	if g.Command != "gemini" {
		t.Fatalf("command = %q", g.Command)
	}

	if _, err := Get("unknown"); err == nil {
		t.Fatal("expected unsupported agent error")
	}
}
