# Regenerate with: vhs docs/demo.tape (from the repo root)
# Requires: vhs, devx in ~/go/bin, tmux
# The recording is fully isolated: its own config, projects dir, and tmux
# server (TMUX_TMPDIR), so it never touches real sessions or config.

Output docs/demo.gif

Set Shell "bash"
Set FontFamily "Meslo LG L for Powerline"
Set FontSize 17
Set Width 1200
Set Height 640
Set Padding 16
Set TypingSpeed 40ms
Set Theme "Catppuccin Mocha"

# --- hidden setup: isolated config, sample projects, styled demo session ---
Hide
Type `alias devx='$HOME/go/bin/devx --config /tmp/devx-demo/config.json'`
Enter
Type "export TMUX_TMPDIR=/tmp/devx-demo-tmux"
Enter
Type "rm -rf /tmp/devx-demo /tmp/devx-demo-tmux && mkdir -p /tmp/devx-demo-tmux /tmp/devx-demo/projects/{blog-engine,api-server,data-pipeline,demo}"
Enter
Type "devx config --default-dir /tmp/devx-demo/projects --default-agent claude"
Enter
Type "devx add blog-engine /tmp/devx-demo/projects/blog-engine --agent claude"
Enter
Type "devx add api-server /tmp/devx-demo/projects/api-server --agent codex"
Enter
Type "devx add data-pipeline /tmp/devx-demo/projects/data-pipeline --agent gemini"
Enter
Type `tmux new-session -d -s demo -n demo -c /tmp/devx-demo/projects/demo 'env PS1="\$ " bash --norc'`
Enter
Type "tmux set -g automatic-rename off && tmux set -g status-right ' devx '"
Enter
Type "clear"
Enter
Sleep 500ms
Show

# --- create a project: registered, git-initialized, and opened in tmux ---
Type "devx create demo --agent none --yes"
Sleep 500ms
Enter
Sleep 2.5s

# --- work happens inside the persistent session ---
Type `echo "your AI agent keeps running in here"`
Sleep 300ms
Enter
Sleep 2s

# --- detach: the session stays alive ---
Ctrl+B
Sleep 300ms
Type "d"
Sleep 1.5s

Type "devx list"
Sleep 500ms
Enter
Sleep 2.5s

Type "devx status"
Sleep 500ms
Enter
Sleep 2.5s

# --- fuzzy-reconnect from anywhere: "dm" resolves to "demo", same state ---
Type "devx dm"
Sleep 500ms
Enter
Sleep 3s

Ctrl+B
Sleep 300ms
Type "d"
Sleep 1.5s

Type "devx stop demo"
Sleep 500ms
Enter
Sleep 2.5s

# --- hidden cleanup ---
Hide
Type "tmux kill-server 2>/dev/null; rm -rf /tmp/devx-demo /tmp/devx-demo-tmux"
Enter
