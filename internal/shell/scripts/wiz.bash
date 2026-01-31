#!/usr/bin/env bash
# wiz shell integration for bash
# Add to .bashrc: eval "$(wiz init bash)"

wiz() {
  if [[ "$1" == "enter" ]]; then
    local output
    output="$(command wiz enter "${@:2}")" || return $?
    eval "$output"
  else
    command wiz "$@"
  fi
}

# Cache variables for status optimization.
_wiz_last_index_mtime=""
_wiz_cached_status=""

_wiz_prompt_hook() {
  if [[ -n "$WIZ_CTX" ]]; then
    # Only call wiz status if the git index has changed.
    local index_file="${WIZ_DIR}/.git/index"
    [[ ! -f "$index_file" ]] && index_file="${WIZ_DIR}/../.git/worktrees/${WIZ_CTX}/index"
    local current_mtime=""
    if [[ -f "$index_file" ]]; then
      current_mtime="$(command stat -f %m "$index_file" 2>/dev/null || command stat -c %Y "$index_file" 2>/dev/null)"
    fi

    if [[ "$current_mtime" != "$_wiz_last_index_mtime" || -z "$_wiz_cached_status" ]]; then
      _wiz_cached_status="$(command wiz status --porcelain 2>/dev/null)"
      _wiz_last_index_mtime="$current_mtime"
    fi

    if [[ -n "$_wiz_cached_status" ]]; then
      local ctx repo branch state ahead behind
      read -r ctx repo branch state ahead behind <<< "$_wiz_cached_status"
      local dirty=""
      [[ "$state" == "dirty" ]] && dirty="*"
      # Terminal title
      printf '\033]0;🧙 %s — %s\007' "$ctx" "$repo"
      # Prompt prefix
      export WIZ_PROMPT="🧙 ${ctx}${dirty}"
    fi
  fi
}

if [[ -z "$WIZ_NO_PROMPT" ]]; then
  PROMPT_COMMAND="_wiz_prompt_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
  # Prepend wiz context to PS1 if active
  _wiz_original_ps1="$PS1"
  _wiz_ps1() {
    if [[ -n "$WIZ_PROMPT" ]]; then
      PS1="\[\033[1;35m\]${WIZ_PROMPT}\[\033[0m\] ${_wiz_original_ps1}"
    fi
  }
  PROMPT_COMMAND="_wiz_ps1${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
