# goreview

Local-first code review tool for GitHub and GitLab. Review diffs, draft comments offline, and publish when ready — all from your terminal.

## Why

Reviewing code on GitHub/GitLab's web UI is slow, disconnected from your local workflow, and impossible offline. goreview lets you review diffs locally, take your time writing comments, and push them to the platform when you're done.

Everything stays local until you explicitly `goreview push`.

## Features

- **Dual mode** — interactive TUI for deep review sessions, composable CLI for scripting
- **Offline-first** — review branches, write comments, and manage sessions without network access
- **Local draft comments** — add, edit, delete comments freely before publishing
- **Multi-line comments** — comment on single lines or ranges
- **Comment types** — `comment`, `blocking`, `nitpick`, `praise`, `question`
- **Git worktree checkout** — review PR branches without leaving your current work
- **GitHub + GitLab** — auto-detects platform from git remote, supports both APIs
- **Walk-through mode** — guided file-by-file review with interactive prompts
- **Syntax highlighting** — diffs rendered with language-aware highlighting
- **Pipeable output** — JSON output, TTY-aware colors, pager support

## Install

```bash
go install github.com/claudioluciano/goreview/cmd/goreview@latest
```

Or build from source:

```bash
git clone https://github.com/claudioluciano/goreview.git
cd goreview
go build -o goreview ./cmd/goreview/
```

## Quick Start

```bash
# Authenticate (uses gh/glab token if available, or set manually)
goreview auth github --token ghp_xxxxx

# List open PRs
goreview list

# Start reviewing a PR
goreview review 42

# Or review any two branches (no PR needed)
goreview review main..feature-branch

# View the diff
goreview diff
goreview diff --stat
goreview diff src/handler.go

# Add comments
goreview comment src/handler.go:14 "Should return a JSON error body"
goreview comment src/handler.go:20-25 "Extract this into a helper"
goreview comment src/auth.go:27 -t blocking "Missing rate limit check"

# Review your draft
goreview comments

# Publish to GitHub/GitLab
goreview push --approve
```

## Commands

| Command | Description |
|---|---|
| `goreview list` | List open PRs/MRs |
| `goreview review <target>` | Start or resume a review (PR number or `base..head`) |
| `goreview diff` | Show diff for the active review |
| `goreview comment <file>:<line> <body>` | Add a comment |
| `goreview comments` | List all draft comments |
| `goreview reviews` | List local review sessions |
| `goreview checkout <PR>` | Checkout PR branch via git worktree |
| `goreview push` | Publish review to GitHub/GitLab |
| `goreview auth <platform>` | Configure authentication |

## Review Workflow

```
goreview review 42          # create local session
goreview diff               # read the changes
goreview comment ...        # draft comments (offline, local)
goreview comment ...        # keep drafting
goreview comments           # review what you wrote
goreview push --approve     # publish everything at once
```

All review data lives in `.goreview/` inside your repo (auto-added to `.gitignore`). You can close the terminal, come back days later, and pick up where you left off with `goreview review --resume pr-42`.

## Walk-through Mode

For a guided review experience:

```bash
goreview review 42 --walk
```

This shows each file's diff one at a time and prompts you to comment or skip after each file.

## Git Worktree Checkout

Review a PR without switching branches:

```bash
goreview checkout 42        # creates .goreview/worktrees/pr-42/
cd .goreview/worktrees/pr-42/
# explore the code, run tests, etc.

goreview checkout --clean 42  # remove when done
```

## Authentication

goreview resolves auth tokens in this order:

1. Stored token from `goreview auth` (`~/.config/goreview/auth.yaml`)
2. Existing CLI token (`gh auth token` / `glab auth token`)
3. Environment variable (`GITHUB_TOKEN` / `GITLAB_TOKEN`)

Most developers using `gh` or `glab` need zero configuration.

## Network Usage

Only these commands touch the network:

| Command | Network |
|---|---|
| `goreview auth` | Yes |
| `goreview list` | Yes |
| `goreview review <PR number>` | Yes (fetches PR metadata) |
| `goreview checkout` | Yes (fetches branch) |
| `goreview push` | Yes |
| Everything else | No |

## Configuration

### Repo-level (`.goreview/config.yaml`)

```yaml
platform: github    # or gitlab (auto-detected from remote if omitted)
remote: origin      # git remote to use
```

### Global (`~/.config/goreview/`)

- `auth.yaml` — stored authentication tokens

## License

MIT
