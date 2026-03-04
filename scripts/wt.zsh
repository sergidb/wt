# wt - Git worktree manager shell wrapper
# Source this file in your .zshrc: source /path/to/wt.zsh
# Requires wt-bin to be in PATH.

wt() {
  # Commands that may return a cd path need stdout captured.
  # Everything else runs directly (pass-through).
  local cmd="${1:-}"

  case "$cmd" in
    run|config|ls|new|rm|prune|help|--help|-h)
      command wt-bin "$@"
      return $?
      ;;
  esac

  # For cd, home, and interactive mode (no args): capture stdout
  local output exit_code
  output=$(command wt-bin "$@")
  exit_code=$?

  if [[ $exit_code -ne 0 ]]; then
    echo "$output"
    return $exit_code
  fi

  # Check if the output signals a cd request
  if [[ "$output" == __WT_CD__:* ]]; then
    local target="${output#__WT_CD__:}"
    cd "$target" || return 1
  else
    # Normal display output
    [[ -n "$output" ]] && echo "$output"
  fi
}
