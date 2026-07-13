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
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/gunwooko/devx/internal/agent"
	"github.com/gunwooko/devx/internal/config"
	"github.com/gunwooko/devx/internal/tmux"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?$`)

type CreateOptions struct {
	ConfigPath string
	Name       string
	Path       string
	Agent      string
	InitGit    bool
	Open       bool
	AssumeYes  bool
	Output     io.Writer
}

type AddOptions struct {
	ConfigPath string
	Name       string
	Path       string
	Agent      string
	Output     io.Writer
}

type ImportOptions struct {
	ConfigPath string
	Dir        string
	Agent      string
	DryRun     bool
	Output     io.Writer
}

type OpenOptions struct {
	ConfigPath    string
	Name          string
	AgentOverride string
	Output        io.Writer
}

type ConfigureOptions struct {
	ConfigPath   string
	DefaultDir   string
	DefaultAgent string
	Output       io.Writer
}

func validateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid project name %q: use letters, numbers, dots, underscores, and hyphens; start and end with a letter or number", name)
	}
	return nil
}

func output(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

// agentFor resolves an agent id against the built-in agents first, then the
// custom agents declared in the config file.
func agentFor(cfg *config.Config, id string) (agent.Agent, error) {
	builtin, err := agent.Get(id)
	if err == nil {
		return builtin, nil
	}
	key := strings.ToLower(strings.TrimSpace(id))
	if custom, ok := cfg.CustomAgents[key]; ok {
		return agent.FromCustom(key, custom.Name, custom.Command)
	}
	return agent.Agent{}, fmt.Errorf("unsupported agent %q (choose: %s)", id, strings.Join(agentIDs(cfg), ", "))
}

// agentIDs returns built-in and custom agent ids, sorted.
func agentIDs(cfg *config.Config) []string {
	ids := agent.IDs()
	for id := range cfg.CustomAgents {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func resolveAgent(cfg *config.Config, requested string, interactive bool) (agent.Agent, error) {
	id := requested
	if id == "" {
		id = cfg.DefaultAgent
	}
	if interactive && requested == "" {
		selected, err := promptAgent(cfg, id)
		if err != nil {
			return agent.Agent{}, err
		}
		id = selected
	}
	return agentFor(cfg, id)
}

func promptAgent(cfg *config.Config, defaultID string) (string, error) {
	if stat, err := os.Stdin.Stat(); err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		return defaultID, nil
	}

	fmt.Printf("Select AI agent [%s] (%s): ", strings.Join(agentIDs(cfg), "/"), defaultID)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultID, nil
	}
	if _, err := agentFor(cfg, line); err != nil {
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

	selected, err := resolveAgent(cfg, opts.Agent, !opts.AssumeYes)
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

	fmt.Fprintf(output(opts.Output), "Created %s\nPath: %s\nAgent: %s\n", opts.Name, projectPath, selected.Name)

	if opts.Open {
		return OpenProject(OpenOptions{ConfigPath: cfgPath, Name: opts.Name, Output: opts.Output})
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

	selected, err := resolveAgent(cfg, opts.Agent, false)
	if err != nil {
		return err
	}
	cfg.Projects[opts.Name] = config.Project{Path: path, Agent: selected.ID}
	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}
	fmt.Fprintf(output(opts.Output), "Added %s\nPath: %s\nAgent: %s\n", opts.Name, path, selected.Name)
	return nil
}

// ImportProjects registers every immediate subdirectory of opts.Dir as a
// project. Hidden directories, invalid names, and names or paths that are
// already registered are skipped with a note.
func ImportProjects(opts ImportOptions) error {
	cfg, cfgPath, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	dir, err := config.ExpandPath(opts.Dir)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read import directory: %w", err)
	}

	selected, err := resolveAgent(cfg, opts.Agent, false)
	if err != nil {
		return err
	}

	registeredPaths := make(map[string]string, len(cfg.Projects))
	for name, p := range cfg.Projects {
		registeredPaths[p.Path] = name
	}

	out := output(opts.Output)
	imported := 0
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		path := filepath.Join(dir, name)
		if err := validateName(name); err != nil {
			fmt.Fprintf(out, "Skipped %s: invalid project name\n", name)
			continue
		}
		if _, exists := cfg.Projects[name]; exists {
			fmt.Fprintf(out, "Skipped %s: name already registered\n", name)
			continue
		}
		if existing, ok := registeredPaths[path]; ok {
			fmt.Fprintf(out, "Skipped %s: path already registered as %q\n", name, existing)
			continue
		}
		if opts.DryRun {
			fmt.Fprintf(out, "Would import %s (%s)\n", name, path)
		} else {
			cfg.Projects[name] = config.Project{Path: path, Agent: selected.ID}
			fmt.Fprintf(out, "Imported %s (%s)\n", name, path)
		}
		imported++
	}

	if opts.DryRun {
		fmt.Fprintf(out, "Dry run: %d project(s) would be imported with agent %s\n", imported, selected.Name)
		return nil
	}
	if imported > 0 {
		if err := config.Save(cfgPath, cfg); err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "Imported %d project(s) with agent %s\n", imported, selected.Name)
	return nil
}

func OpenProject(opts OpenOptions) error {
	cfg, _, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	name := opts.Name
	project, ok := cfg.Projects[name]
	if !ok {
		name, err = resolveProject(cfg, opts.Name)
		if err != nil {
			return err
		}
		project = cfg.Projects[name]
		fmt.Fprintf(output(opts.Output), "Opening %s\n", name)
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
	selected, err := agentFor(cfg, agentID)
	if err != nil {
		return err
	}

	session := tmux.SessionName(name)
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
		fmt.Fprintf(output(opts.Output), "Started %s with %s\n", name, selected.Name)
	}
	return tmux.AttachOrSwitch(session)
}

// resolveProject fuzzily matches input against registered project names.
// A unique candidate is used directly; multiple candidates prompt for a
// choice on a terminal and fail otherwise.
func resolveProject(cfg *config.Config, input string) (string, error) {
	candidates := matchProjects(config.Names(cfg), input)
	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("project %q is not registered; run `devx list` or `devx add`", input)
	case 1:
		return candidates[0], nil
	}
	return promptProject(input, candidates)
}

// matchProjects returns the best non-empty tier of matches for input:
// case-insensitive prefix, then substring, then in-order subsequence
// (e.g. "nls" matches "novel-love-story").
func matchProjects(names []string, input string) []string {
	in := strings.ToLower(input)
	if in == "" {
		return nil
	}
	var prefix, substr, subseq []string
	for _, name := range names {
		n := strings.ToLower(name)
		switch {
		case strings.HasPrefix(n, in):
			prefix = append(prefix, name)
		case strings.Contains(n, in):
			substr = append(substr, name)
		case isSubsequence(in, n):
			subseq = append(subseq, name)
		}
	}
	for _, tier := range [][]string{prefix, substr, subseq} {
		if len(tier) > 0 {
			sort.Strings(tier)
			return tier
		}
	}
	return nil
}

func isSubsequence(needle, haystack string) bool {
	if needle == "" {
		return false
	}
	i := 0
	for _, r := range haystack {
		if i < len(needle) && rune(needle[i]) == r {
			i++
		}
	}
	return i == len(needle)
}

func promptProject(input string, candidates []string) (string, error) {
	if stat, err := os.Stdin.Stat(); err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		return "", fmt.Errorf("%q matches multiple projects: %s", input, strings.Join(candidates, ", "))
	}
	fmt.Printf("%q matches multiple projects:\n", input)
	for i, name := range candidates {
		fmt.Printf("  %d) %s\n", i+1, name)
	}
	fmt.Print("Select [1]: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", fmt.Errorf("%q matches multiple projects: %s", input, strings.Join(candidates, ", "))
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return candidates[0], nil
	}
	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > len(candidates) {
		return "", fmt.Errorf("invalid selection %q", line)
	}
	return candidates[idx-1], nil
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

func StopProject(configPath, name string, out io.Writer) error {
	cfg, _, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q is not registered", name)
	}
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
	cfg, path, err := config.Load(configPath)
	if err != nil {
		return err
	}
	selected, err := agentFor(cfg, agentID)
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
		a, err := agentFor(cfg, opts.DefaultAgent)
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
		commandResult("gemini", false),
		commandResult("opencode", false),
		commandResult("tailscale", false),
	}

	cfg, path, cfgErr := config.Load(configPath)
	results = append(results, doctorResult{name: "config", ok: cfgErr == nil, info: path})
	if cfgErr == nil {
		customIDs := make([]string, 0, len(cfg.CustomAgents))
		for id := range cfg.CustomAgents {
			customIDs = append(customIDs, id)
		}
		sort.Strings(customIDs)
		for _, id := range customIDs {
			c := cfg.CustomAgents[id]
			a, err := agent.FromCustom(id, c.Name, c.Command)
			if err != nil {
				results = append(results, doctorResult{name: "custom agent " + id, ok: false, info: err.Error()})
				continue
			}
			r := commandResult(agent.Executable(a), false)
			r.name = "custom agent " + id
			results = append(results, r)
		}
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
