# Project Guidelines

## Code Quality
- Follow Go best practices and idiomatic patterns
- Keep code clean, well-structured, and maintainable

## UX & Interface
- The TUI must provide an excellent user experience: professional, polished, and visually appealing
- Use lipgloss/bubbletea to create smooth, intuitive interactions
- Prioritize clarity and ease of use in all output

## Workflow
- Always finish any implementation with a proposal on how to verify the changes (e.g. commands to run, manual checks, etc.)
- If there is nothing to verify, suggest what to work on next

## Git & Commits
- Never sign commits or PRs as Claude or with Claude's identity
- Never use Co-Authored-By headers referencing Claude
- Use the repository's existing git user config for commits
- Do not use --no-gpg-sign or override user/email in commit commands
