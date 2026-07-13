package agent

import (
	"fmt"
	"os/exec"
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
	"gemini": {ID: "gemini", Name: "Gemini CLI", Command: "gemini"},
	"none":   {ID: "none", Name: "Shell only", Command: ""},
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
	_, err := exec.LookPath(a.Command)
	return err == nil
}
