# wt - Git worktree manager shell wrapper
# Source this file in your .zshrc: source /path/to/wt.zsh
# Requires wt-bin to be in PATH.

wt() {
  local output exit_code

  # Capture stdout (result/cd signal), stderr goes to terminal (TUI rendering).
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
