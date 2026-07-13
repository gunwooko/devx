package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gunwooko/devx/internal/config"
)

func testConfigPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "config.json")
}

func mustLoad(t *testing.T, path string) *config.Config {
	t.Helper()
	cfg, _, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestValidateName(t *testing.T) {
	valid := []string{"demo", "my.app", "api_service", "a", "web-2", "A1"}
	for _, name := range valid {
		if err := validateName(name); err != nil {
			t.Errorf("validateName(%q) = %v, want nil", name, err)
		}
	}
	invalid := []string{"", "-demo", "demo-", ".demo", "demo.", "my app", "a/b", "한글"}
	for _, name := range invalid {
		if err := validateName(name); err == nil {
			t.Errorf("validateName(%q) = nil, want error", name)
		}
	}
}

func TestCreateProject(t *testing.T) {
	cfgPath := testConfigPath(t)
	projectPath := filepath.Join(t.TempDir(), "demo")
	var out bytes.Buffer

	err := CreateProject(CreateOptions{
		ConfigPath: cfgPath,
		Name:       "demo",
		Path:       projectPath,
		Agent:      "none",
		AssumeYes:  true,
		Output:     &out,
	})
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(projectPath)
	if err != nil || !info.IsDir() {
		t.Fatalf("project directory was not created: %v", err)
	}
	cfg := mustLoad(t, cfgPath)
	project, ok := cfg.Projects["demo"]
	if !ok {
		t.Fatal("project was not registered")
	}
	if project.Agent != "none" {
		t.Fatalf("agent = %q, want none", project.Agent)
	}
	if !strings.Contains(out.String(), "Created demo") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCreateProjectInitGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	cfgPath := testConfigPath(t)
	projectPath := filepath.Join(t.TempDir(), "demo")

	err := CreateProject(CreateOptions{
		ConfigPath: cfgPath,
		Name:       "demo",
		Path:       projectPath,
		Agent:      "none",
		InitGit:    true,
		AssumeYes:  true,
		Output:     &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !isGitRepo(projectPath) {
		t.Fatal("git repository was not initialized")
	}
}

func TestCreateProjectDuplicate(t *testing.T) {
	cfgPath := testConfigPath(t)
	opts := CreateOptions{
		ConfigPath: cfgPath,
		Name:       "demo",
		Path:       filepath.Join(t.TempDir(), "demo"),
		Agent:      "none",
		AssumeYes:  true,
		Output:     &bytes.Buffer{},
	}
	if err := CreateProject(opts); err != nil {
		t.Fatal(err)
	}
	if err := CreateProject(opts); err == nil {
		t.Fatal("expected error for duplicate project")
	}
}

func TestCreateProjectInvalidName(t *testing.T) {
	err := CreateProject(CreateOptions{ConfigPath: testConfigPath(t), Name: "bad name", AssumeYes: true})
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestCreateProjectPathIsFile(t *testing.T) {
	cfgPath := testConfigPath(t)
	file := filepath.Join(t.TempDir(), "occupied")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := CreateProject(CreateOptions{
		ConfigPath: cfgPath,
		Name:       "demo",
		Path:       file,
		Agent:      "none",
		AssumeYes:  true,
		Output:     &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error when path exists and is not a directory")
	}
}

func TestAddProject(t *testing.T) {
	cfgPath := testConfigPath(t)
	dir := t.TempDir()
	var out bytes.Buffer

	err := AddProject(AddOptions{
		ConfigPath: cfgPath,
		Name:       "existing",
		Path:       dir,
		Agent:      "codex",
		Output:     &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := mustLoad(t, cfgPath)
	if cfg.Projects["existing"].Agent != "codex" {
		t.Fatalf("agent = %q, want codex", cfg.Projects["existing"].Agent)
	}
}

func TestAddProjectMissingPath(t *testing.T) {
	err := AddProject(AddOptions{
		ConfigPath: testConfigPath(t),
		Name:       "ghost",
		Path:       filepath.Join(t.TempDir(), "missing"),
		Agent:      "none",
		Output:     &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestOpenProjectUnregistered(t *testing.T) {
	err := OpenProject(OpenOptions{ConfigPath: testConfigPath(t), Name: "ghost", Output: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("err = %v, want not-registered error", err)
	}
}

func TestMatchProjects(t *testing.T) {
	names := []string{"novel-love-story", "my-app", "app-server", "raon"}
	cases := []struct {
		input string
		want  []string
	}{
		{"novel", []string{"novel-love-story"}},  // prefix
		{"NOVEL", []string{"novel-love-story"}},  // case-insensitive
		{"app", []string{"app-server"}},          // prefix tier beats substring ("my-app")
		{"pp", []string{"app-server", "my-app"}}, // substring, sorted
		{"nls", []string{"novel-love-story"}},    // subsequence
		{"raon", []string{"raon"}},               // exact is also a prefix
		{"zzz", nil},                             // no match
		{"", nil},                                // empty input matches nothing useful
	}
	for _, c := range cases {
		got := matchProjects(names, c.input)
		if len(got) != len(c.want) {
			t.Errorf("matchProjects(%q) = %v, want %v", c.input, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("matchProjects(%q) = %v, want %v", c.input, got, c.want)
				break
			}
		}
	}
}

func TestIsSubsequence(t *testing.T) {
	if !isSubsequence("nls", "novel-love-story") {
		t.Error("nls should be a subsequence of novel-love-story")
	}
	if isSubsequence("nsl", "novel-love-story") {
		t.Error("nsl is out of order and must not match")
	}
	if isSubsequence("", "anything") {
		t.Error("empty needle must not match")
	}
}

func TestResolveProject(t *testing.T) {
	cfg := &config.Config{Projects: map[string]config.Project{
		"novel-love-story": {},
		"my-app":           {},
		"app-server":       {},
	}}

	name, err := resolveProject(cfg, "novel")
	if err != nil || name != "novel-love-story" {
		t.Fatalf("resolveProject(novel) = %q, %v", name, err)
	}

	if _, err := resolveProject(cfg, "zzz"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("err = %v, want not-registered error", err)
	}

	// Ambiguous input without a terminal reports the candidates.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	if _, err := resolveProject(cfg, "pp"); err == nil || !strings.Contains(err.Error(), "matches multiple projects") {
		t.Fatalf("err = %v, want ambiguous-match error", err)
	}
}

func TestSetAgent(t *testing.T) {
	cfgPath := testConfigPath(t)
	dir := t.TempDir()
	if err := AddProject(AddOptions{ConfigPath: cfgPath, Name: "demo", Path: dir, Agent: "claude", Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := SetAgent(cfgPath, "demo", "codex", &out); err != nil {
		t.Fatal(err)
	}
	cfg := mustLoad(t, cfgPath)
	if cfg.Projects["demo"].Agent != "codex" {
		t.Fatalf("agent = %q, want codex", cfg.Projects["demo"].Agent)
	}

	if err := SetAgent(cfgPath, "ghost", "codex", &out); err == nil {
		t.Fatal("expected error for unregistered project")
	}
	if err := SetAgent(cfgPath, "demo", "bogus", &out); err == nil {
		t.Fatal("expected error for unsupported agent")
	}
}

func TestRemoveProject(t *testing.T) {
	cfgPath := testConfigPath(t)
	dir := t.TempDir()
	if err := AddProject(AddOptions{ConfigPath: cfgPath, Name: "devx-test-remove", Path: dir, Agent: "none", Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := RemoveProject(cfgPath, "devx-test-remove", false, &out); err != nil {
		t.Fatal(err)
	}
	cfg := mustLoad(t, cfgPath)
	if _, ok := cfg.Projects["devx-test-remove"]; ok {
		t.Fatal("project was not removed")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("project files must not be deleted on remove")
	}

	if err := RemoveProject(cfgPath, "ghost", false, &out); err == nil {
		t.Fatal("expected error for unregistered project")
	}
}

func TestStopProjectUnregistered(t *testing.T) {
	err := StopProject(testConfigPath(t), "ghost", &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("err = %v, want not-registered error", err)
	}
}

func TestStopProjectNoSession(t *testing.T) {
	cfgPath := testConfigPath(t)
	if err := AddProject(AddOptions{ConfigPath: cfgPath, Name: "devx-test-stop", Path: t.TempDir(), Agent: "none", Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := StopProject(cfgPath, "devx-test-stop", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No active session") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestListProjects(t *testing.T) {
	cfgPath := testConfigPath(t)
	if err := AddProject(AddOptions{ConfigPath: cfgPath, Name: "beta", Path: t.TempDir(), Agent: "none", Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	if err := AddProject(AddOptions{ConfigPath: cfgPath, Name: "alpha", Path: t.TempDir(), Agent: "claude", Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := ListProjects(cfgPath, &out); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "NAME") || !strings.Contains(got, "alpha") || !strings.Contains(got, "beta") {
		t.Fatalf("output = %q", got)
	}
	if strings.Index(got, "alpha") > strings.Index(got, "beta") {
		t.Fatalf("projects are not sorted: %q", got)
	}
}

func TestStatus(t *testing.T) {
	cfgPath := testConfigPath(t)
	if err := AddProject(AddOptions{ConfigPath: cfgPath, Name: "devx-test-status", Path: t.TempDir(), Agent: "none", Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Status(cfgPath, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "devx-test-status") || !strings.Contains(out.String(), "stopped") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestConfigure(t *testing.T) {
	cfgPath := testConfigPath(t)
	dir := t.TempDir()
	var out bytes.Buffer

	err := Configure(ConfigureOptions{
		ConfigPath:   cfgPath,
		DefaultDir:   dir,
		DefaultAgent: "codex",
		Output:       &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := mustLoad(t, cfgPath)
	if cfg.DefaultProjectsDir != dir {
		t.Fatalf("default dir = %q, want %q", cfg.DefaultProjectsDir, dir)
	}
	if cfg.DefaultAgent != "codex" {
		t.Fatalf("default agent = %q, want codex", cfg.DefaultAgent)
	}

	if err := Configure(ConfigureOptions{ConfigPath: cfgPath, DefaultAgent: "bogus", Output: &out}); err == nil {
		t.Fatal("expected error for unsupported default agent")
	}
}
