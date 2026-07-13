package agent

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

type Agent struct {
	ID      string
	Name    string
	Command string
}

var supported = map[string]Agent{
	"claude": {ID: "claude", Name: "Claude Code", Command: "claude"},
	"codex":  {ID: "codex", Name: "Codex CLI", Command: "codex"},
	"gemini":   {ID: "gemini", Name: "Gemini CLI", Command: "gemini"},
	"opencode": {ID: "opencode", Name: "OpenCode", Command: "opencode"},
	"none":     {ID: "none", Name: "Shell only", Command: ""},
}

var validCommandToken = regexp.MustCompile(`^[A-Za-z0-9._/=-]+$`)

// FromCustom builds an Agent from a user-defined config entry. The command
// is restricted to a plain executable with simple arguments so config values
// can never smuggle shell control characters into the tmux session command.
func FromCustom(id, name, command string) (Agent, error) {
	if _, exists := supported[id]; exists {
		return Agent{}, fmt.Errorf("custom agent %q shadows a built-in agent", id)
	}
	tokens := strings.Fields(command)
	if len(tokens) == 0 {
		return Agent{}, fmt.Errorf("custom agent %q has an empty command", id)
	}
	for _, token := range tokens {
		if !validCommandToken.MatchString(token) {
			return Agent{}, fmt.Errorf("custom agent %q: unsupported characters in command token %q", id, token)
		}
	}
	if name == "" {
		name = id
	}
	return Agent{ID: id, Name: name, Command: strings.Join(tokens, " ")}, nil
}

func Get(id string) (Agent, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	a, ok := supported[id]
	if !ok {
		return Agent{}, fmt.Errorf("unsupported agent %q (choose: %s)", id, strings.Join(IDs(), ", "))
	}
	return a, nil
}

func IDs() []string {
	ids := make([]string, 0, len(supported))
	for id := range supported {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func Installed(a Agent) bool {
	if a.Command == "" {
		return true
	}
	_, err := exec.LookPath(Executable(a))
	return err == nil
}

// Executable returns the program to look up in PATH: the first token of the
// command, since custom agents may carry arguments.
func Executable(a Agent) string {
	fields := strings.Fields(a.Command)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
