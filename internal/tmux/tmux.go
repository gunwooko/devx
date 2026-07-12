package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var invalidSessionChars = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

func SessionName(name string) string {
	cleaned := invalidSessionChars.ReplaceAllString(name, "-")
	cleaned = strings.Trim(cleaned, "-.")
	if cleaned == "" {
		return "devx"
	}
	return cleaned
}

func Installed() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func Exists(session string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", "="+session)
	return cmd.Run() == nil
}

func CreateDetached(session, directory, shellCommand string) error {
	args := []string{"new-session", "-d", "-s", session, "-c", directory}
	if shellCommand != "" {
		args = append(args, shellCommand)
	}
	cmd := exec.Command("tmux", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create tmux session: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func AttachOrSwitch(session string) error {
	var cmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		cmd = exec.Command("tmux", "switch-client", "-t", session)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", session)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("attach tmux session: %w", err)
	}
	return nil
}

func Kill(session string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", "="+session)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("stop tmux session: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
