# wt

A fast terminal tool for navigating and managing git worktrees. Built for developers who use Claude Code or multiple worktrees and need to quickly switch between them, run services, and keep things tidy.

## Why

Working with git worktrees means constantly `cd`-ing into long paths like `.claude/worktrees/fix-auth`. When you need to run backend and frontend in separate terminals for manual testing, it gets tedious fast. `wt` fixes that.

## Install

```bash
# Build
go build -o wt-bin .

# Install binary
sudo make install
# or: make install INSTALL_DIR=~/.local/bin

# Add shell wrapper to .zshrc
echo 'source /path/to/wt/scripts/wt.zsh' >> ~/.zshrc
source ~/.zshrc
```

> The shell wrapper is required — it enables `wt cd` and `wt home` to change your shell's directory.

## Usage

### Interactive mode

```bash
wt
```

Arrow keys to browse worktrees, enter to select, then pick an action (cd, run, remove, etc). Press `c` to configure services.

### Commands

```bash
wt ls                  # List all worktrees
wt cd <name>           # Navigate to a worktree
wt home                # Go back to the main repo root
wt new <branch>        # Create a worktree in .claude/worktrees/
wt rm <name>           # Remove a worktree
wt prune               # Clean up stale worktree entries
wt run [name] [svc..]  # Run services (all or specific ones)
wt config              # Configure services interactively
```

Name matching is fuzzy — `wt cd fix` matches `fix-auth` if it's the only match.

### Service runner

Define services in `.wt.yaml` at your repo root:

```yaml
services:
  backend:
    cmd: "go run ./cmd/server"
    dir: "."
    color: "green"
  frontend:
    cmd: "npm run dev"
    dir: "./frontend"
    color: "cyan"
```

Then:

```bash
wt run fix-auth           # Run all services for that worktree
wt run fix-auth backend   # Run only the backend
wt run                    # Run services for the current worktree
```

Output is docker-compose style with colored `[service]` prefixes. `Ctrl+C` stops all.

You can also create this config interactively with `wt config`.

## License

MIT
