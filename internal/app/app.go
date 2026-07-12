package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/gunwooko/devx/internal/agent"
	"github.com/gunwooko/devx/internal/config"
	"github.com/gunwooko/devx/internal/tmux"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type CreateOptions struct {
	ConfigPath string
	Name       string
	Path       string
	Agent      string
	InitGit    bool
	Open       bool
	AssumeYes  bool
}

type AddOptions struct {
	ConfigPath string
	Name       string
	Path       string
	Agent      string
}

type OpenOptions struct {
	ConfigPath    string
	Name          string
	AgentOverride string
}

type ConfigureOptions struct {
	ConfigPath   string
	DefaultDir   string
	DefaultAgent string
	Output       io.Writer
}

func validateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid project name %q: use letters, numbers, dots, underscores, and hyphens", name)
	}
	return nil
}

func resolveAgent(requested, fallback string, interactive bool) (agent.Agent, error) {
	id := requested
	if id == "" {
		id = fallback
	}
	if interactive && requested == "" {
		selected, err := promptAgent(id)
		if err != nil {
			return agent.Agent{}, err
		}
		id = selected
	}
	return agent.Get(id)
}

func promptAgent(defaultID string) (string, error) {
	if stat, err := os.Stdin.Stat(); err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		return defaultID, nil
	}

	fmt.Printf("Select AI agent [claude/codex/none] (%s): ", defaultID)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultID, nil
	}
	if _, err := agent.Get(line); err != nil {
		return "", err
	}
	return line, nil
}

func CreateProject(opts CreateOptions) error {
	if err := validateName(opts.Name); err != nil {
		return err
	}

	cfg, cfgPath, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	if _, exists := cfg.Projects[opts.Name]; exists {
		return fmt.Errorf("project %q is already registered", opts.Name)
	}

	projectPath := opts.Path
	if projectPath == "" {
		projectPath = filepath.Join(cfg.DefaultProjectsDir, opts.Name)
	}
	projectPath, err = config.ExpandPath(projectPath)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	selected, err := resolveAgent(opts.Agent, cfg.DefaultAgent, !opts.AssumeYes)
	if err != nil {
		return err
	}

	info, statErr := os.Stat(projectPath)
	if statErr == nil && !info.IsDir() {
		return fmt.Errorf("path exists and is not a directory: %s", projectPath)
	}
	if statErr != nil && !os.IsNotExist(statErr) {
		return statErr
	}
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	if opts.InitGit && !isGitRepo(projectPath) {
		if output, err := exec.Command("git", "-C", projectPath, "init").CombinedOutput(); err != nil {
			return fmt.Errorf("initialize git repository: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}

	cfg.Projects[opts.Name] = config.Project{Path: projectPath, Agent: selected.ID}
	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}

	fmt.Printf("Created %s\nPath: %s\nAgent: %s\n", opts.Name, projectPath, selected.Name)

	if opts.Open {
		return OpenProject(OpenOptions{ConfigPath: cfgPath, Name: opts.Name})
	}
	return nil
}

func AddProject(opts AddOptions) error {
	if err := validateName(opts.Name); err != nil {
		return err
	}
	cfg, cfgPath, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	if _, exists := cfg.Projects[opts.Name]; exists {
		return fmt.Errorf("project %q is already registered", opts.Name)
	}

	path, err := config.ExpandPath(opts.Path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("access project directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", path)
	}

	selected, err := resolveAgent(opts.Agent, cfg.DefaultAgent, false)
	if err != nil {
		return err
	}
	cfg.Projects[opts.Name] = config.Project{Path: path, Agent: selected.ID}
	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Added %s\nPath: %s\nAgent: %s\n", opts.Name, path, selected.Name)
	return nil
}

func OpenProject(opts OpenOptions) error {
	cfg, _, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	project, ok := cfg.Projects[opts.Name]
	if !ok {
		return fmt.Errorf("project %q is not registered; run `devx list` or `devx add`", opts.Name)
	}
	if _, err := os.Stat(project.Path); err != nil {
		return fmt.Errorf("project path is unavailable: %w", err)
	}
	if !tmux.Installed() {
		return fmt.Errorf("tmux is not installed")
	}

	agentID := project.Agent
	if opts.AgentOverride != "" {
		agentID = opts.AgentOverride
	}
	selected, err := agent.Get(agentID)
	if err != nil {
		return err
	}

	session := tmux.SessionName(opts.Name)
	if !tmux.Exists(session) {
		if !agent.Installed(selected) {
			return fmt.Errorf("%s command %q was not found in PATH", selected.Name, selected.Command)
		}
		command := ""
		if selected.Command != "" {
			command = selected.Command + "; exec ${SHELL:-/bin/sh}"
		}
		if err := tmux.CreateDetached(session, project.Path, command); err != nil {
			return err
		}
		fmt.Printf("Started %s with %s\n", opts.Name, selected.Name)
	}
	return tmux.AttachOrSwitch(session)
}

func ListProjects(configPath string, out io.Writer) error {
	cfg, _, err := config.Load(configPath)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tAGENT\tPATH")
	for _, name := range config.Names(cfg) {
		project := cfg.Projects[name]
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, project.Agent, project.Path)
	}
	return w.Flush()
}

func Status(configPath string, out io.Writer) error {
	cfg, _, err := config.Load(configPath)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSESSION\tAGENT\tPATH")
	for _, name := range config.Names(cfg) {
		state := "stopped"
		if tmux.Exists(tmux.SessionName(name)) {
			state = "running"
		}
		p := cfg.Projects[name]
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, state, p.Agent, p.Path)
	}
	return w.Flush()
}

func StopProject(name string, out io.Writer) error {
	session := tmux.SessionName(name)
	if !tmux.Exists(session) {
		fmt.Fprintf(out, "No active session for %s\n", name)
		return nil
	}
	if err := tmux.Kill(session); err != nil {
		return err
	}
	fmt.Fprintf(out, "Stopped %s\n", name)
	return nil
}

func SetAgent(configPath, name, agentID string, out io.Writer) error {
	selected, err := agent.Get(agentID)
	if err != nil {
		return err
	}
	cfg, path, err := config.Load(configPath)
	if err != nil {
		return err
	}
	project, ok := cfg.Projects[name]
	if !ok {
		return fmt.Errorf("project %q is not registered", name)
	}
	project.Agent = selected.ID
	cfg.Projects[name] = project
	if err := config.Save(path, cfg); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s now uses %s\n", name, selected.Name)
	return nil
}

func RemoveProject(configPath, name string, force bool, out io.Writer) error {
	cfg, path, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q is not registered", name)
	}
	session := tmux.SessionName(name)
	if tmux.Exists(session) {
		if !force {
			return fmt.Errorf("project %q has an active session; stop it first or use --force", name)
		}
		if err := tmux.Kill(session); err != nil {
			return err
		}
	}
	delete(cfg.Projects, name)
	if err := config.Save(path, cfg); err != nil {
		return err
	}
	fmt.Fprintf(out, "Removed %s from devx; project files were not deleted\n", name)
	return nil
}

func Configure(opts ConfigureOptions) error {
	cfg, path, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	changed := false
	if opts.DefaultDir != "" {
		dir, err := config.ExpandPath(opts.DefaultDir)
		if err != nil {
			return err
		}
		cfg.DefaultProjectsDir = dir
		changed = true
	}
	if opts.DefaultAgent != "" {
		a, err := agent.Get(opts.DefaultAgent)
		if err != nil {
			return err
		}
		cfg.DefaultAgent = a.ID
		changed = true
	}
	if changed {
		if err := config.Save(path, cfg); err != nil {
			return err
		}
	}
	fmt.Fprintf(opts.Output, "Config: %s\nDefault project directory: %s\nDefault agent: %s\n", path, cfg.DefaultProjectsDir, cfg.DefaultAgent)
	return nil
}

type doctorResult struct {
	name string
	ok   bool
	info string
}

func Doctor(configPath string, out io.Writer) error {
	results := []doctorResult{
		commandResult("git", true),
		commandResult("tmux", true),
		commandResult("claude", false),
		commandResult("codex", false),
		commandResult("tailscale", false),
	}

	cfg, path, cfgErr := config.Load(configPath)
	results = append(results, doctorResult{name: "config", ok: cfgErr == nil, info: path})
	if cfgErr == nil {
		info, err := os.Stat(cfg.DefaultProjectsDir)
		ok := err == nil && info.IsDir()
		detail := cfg.DefaultProjectsDir
		if os.IsNotExist(err) {
			detail += " (will be created on first project)"
			ok = true
		}
		results = append(results, doctorResult{name: "default project directory", ok: ok, info: detail})
	}

	requiredFailure := false
	for _, r := range results {
		symbol := "✓"
		if !r.ok {
			symbol = "!"
			if r.name == "git" || r.name == "tmux" || r.name == "config" {
				requiredFailure = true
			}
		}
		fmt.Fprintf(out, "%s %-26s %s\n", symbol, r.name, r.info)
	}
	if requiredFailure {
		return fmt.Errorf("one or more required checks failed")
	}
	return nil
}

func commandResult(name string, required bool) doctorResult {
	path, err := exec.LookPath(name)
	info := path
	if err != nil {
		if required {
			info = "required; not found"
		} else {
			info = "optional; not found"
		}
	}
	return doctorResult{name: name, ok: err == nil, info: info}
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

// SortedAgents is kept small and explicit for predictable CLI output.
func SortedAgents() []string {
	ids := agent.IDs()
	sort.Strings(ids)
	return ids
}
