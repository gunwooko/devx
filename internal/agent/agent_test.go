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

	o, err := Get("opencode")
	if err != nil {
		t.Fatal(err)
	}
	if o.Command != "opencode" {
		t.Fatalf("command = %q", o.Command)
	}

	if _, err := Get("unknown"); err == nil {
		t.Fatal("expected unsupported agent error")
	}
}

func TestFromCustom(t *testing.T) {
	a, err := FromCustom("aider", "", "aider --model gpt-4o")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "aider" {
		t.Fatalf("name = %q, want id fallback", a.Name)
	}
	if a.Command != "aider --model gpt-4o" {
		t.Fatalf("command = %q", a.Command)
	}
	if Executable(a) != "aider" {
		t.Fatalf("executable = %q", Executable(a))
	}

	invalid := []struct {
		id, command string
	}{
		{"claude", "claude"},          // shadows built-in
		{"empty", "   "},              // empty command
		{"inject", "aider; rm -rf /"}, // shell metacharacters
		{"subst", "aider $(whoami)"},
		{"pipe", "aider | tee log"},
	}
	for _, tc := range invalid {
		if _, err := FromCustom(tc.id, "", tc.command); err == nil {
			t.Errorf("FromCustom(%q, %q) = nil, want error", tc.id, tc.command)
		}
	}
}
