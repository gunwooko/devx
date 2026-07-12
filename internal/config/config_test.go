package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		DefaultProjectsDir: filepath.Join(dir, "projects"),
		DefaultAgent:       "codex",
		Projects: map[string]Project{
			"demo": {Path: filepath.Join(dir, "demo"), Agent: "claude"},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, loadedPath, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loadedPath != path {
		t.Fatalf("path = %q, want %q", loadedPath, path)
	}
	if loaded.DefaultAgent != "codex" {
		t.Fatalf("default agent = %q", loaded.DefaultAgent)
	}
	if loaded.Projects["demo"].Agent != "claude" {
		t.Fatalf("project agent = %q", loaded.Projects["demo"].Agent)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %o, want 600", info.Mode().Perm())
	}
}
