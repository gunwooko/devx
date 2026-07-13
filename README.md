# devx

`devx` is a small Go CLI for opening AI coding projects inside persistent `tmux` sessions.

It supports project-level defaults for:

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code)
- [Codex CLI](https://github.com/openai/codex)
- Shell-only sessions

The main workflow is:

```bash
devx create novel-love-story
devx novel-love-story
```

When a project is opened, `devx` creates or reconnects to a tmux session. That makes the same AI CLI session available from a Mac terminal, SSH, Termius, or another tmux client.

## Features

- Create and Git-initialize new projects
- Register existing projects
- Choose Claude Code, Codex, or shell-only per project
- Override the AI agent for a new session
- Create, attach, switch, inspect, and stop tmux sessions
- Safe behavior when invoked from inside tmux
- JSON configuration under the OS user config directory
- Dependency and configuration checks with `devx doctor`

## Requirements

Required:

- Go 1.23 or newer to build
- Git
- tmux

Optional:

- Claude Code (`claude`)
- Codex CLI (`codex`)
- Tailscale for secure remote access

## Install with Homebrew

```bash
brew install gunwooko/tap/devx
```

## Install from source

Clone the repository and build:

```bash
git clone https://github.com/gunwooko/devx.git
cd devx
go mod tidy
go test ./...
go install .
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

For zsh:

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Check the installation:

```bash
devx doctor
```

## Quick start

Set defaults:

```bash
devx config \
  --default-dir ~/Projects/personal \
  --default-agent claude
```

Create a project interactively:

```bash
devx create novel-love-story
```

Create without prompts:

```bash
devx create novel-love-story \
  --agent claude \
  --yes
```

Create in a custom directory:

```bash
devx create raon \
  --path ~/Projects/startup/raon \
  --agent codex
```

Register an existing project:

```bash
devx add novel ~/Projects/personal/novel --agent claude
```

Open it:

```bash
devx novel
```

The explicit equivalent is:

```bash
devx open novel
```

## Commands

| Command | Description |
|---------|-------------|
| `devx <project>` | Shorthand for `devx open <project>` |
| `devx create <name>` | Create a directory, initialize Git, register, and open it |
| `devx add <name> <path>` | Register an existing directory without touching it |
| `devx open <name>` | Start or reattach the project's tmux session |
| `devx list` | List registered projects (name, agent, path) |
| `devx status` | Like `list`, plus whether each session is running or stopped |
| `devx stop <name>` | Kill the project's tmux session; the project stays registered |
| `devx agent <name> <claude\|codex\|none>` | Change the project's default agent |
| `devx remove <name>` | Unregister a project; files are never deleted |
| `devx config` | Show or update global defaults |
| `devx doctor` | Check dependencies and configuration |
| `devx completion <shell>` | Generate shell completion (bash, zsh, fish, powershell) |

### Agent override

Use another agent for a newly created tmux session without changing the project default:

```bash
devx open novel --agent codex
```

If the session is already running, `devx` reconnects to that session. Stop it first to start a new session with a different agent:

```bash
devx stop novel
devx open novel --agent codex
```

### Change the project default

```bash
devx agent novel codex
```

### Remove registration

This never deletes project files:

```bash
devx remove novel
```

Use `--force` to stop an active session before removing:

```bash
devx remove novel --force
```

## Configuration

By default, macOS uses:

```text
~/Library/Application Support/devx/config.json
```

Linux generally uses:

```text
~/.config/devx/config.json
```

Example:

```json
{
  "defaultProjectsDir": "/Users/gunwoo/Projects/personal",
  "defaultAgent": "claude",
  "projects": {
    "novel": {
      "path": "/Users/gunwoo/Projects/personal/novel",
      "agent": "claude"
    },
    "raon": {
      "path": "/Users/gunwoo/Projects/startup/raon",
      "agent": "codex"
    }
  }
}
```

Use a custom configuration file:

```bash
devx --config /path/to/config.json list
```

## Remote workflow

On the Mac:

```bash
devx novel
```

Detach from tmux without stopping the agent:

```text
Ctrl-b d
```

From a phone over SSH:

```bash
devx novel
```

`devx` reconnects to the same tmux session. You do not need to `cd` into the project directory first.

## Development

```bash
go mod tidy
go test ./...
go vet ./...
go build ./...
```

## Security

- `devx` does not manage SSH keys, VPNs, or Tailscale ACLs.
- The config file is written with user-only permissions where supported.
- Project files are never deleted by `devx remove`.
- Agent commands are selected from a fixed allowlist rather than arbitrary shell input.

## Roadmap

- Per-project environment variables
- More agents such as Gemini CLI and OpenCode
- Bulk import of existing project directories
- Fuzzy matching and interactive selection for `devx open`
- Optional TUI project picker

## License

MIT
