#!/usr/bin/env zsh
# wiz shell integration for zsh
# Add to .zshrc: eval "$(wiz init zsh)"

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
      print -Pn "\033]0;🧙 ${ctx} — ${repo}\007"
      # iTerm2 badge (if supported)
      if [[ "$TERM_PROGRAM" == "iTerm.app" ]]; then
        printf "\033]1337;SetBadgeFormat=%s\007" "$(echo -n "🧙 ${ctx}" | base64)"
      fi
      # Export for prompt
      export WIZ_PROMPT="🧙 ${ctx}${dirty}"
    fi
  fi
}

if [[ -z "$WIZ_NO_PROMPT" ]]; then
  precmd_functions+=(_wiz_prompt_hook)
  # Prompt prefix
  _wiz_setup_prompt() {
    if [[ -n "$WIZ_PROMPT" ]]; then
      PROMPT="%F{magenta}${WIZ_PROMPT}%f ${PROMPT}"
    fi
  }
  # Only set once; the hook updates WIZ_PROMPT dynamically.
  precmd_functions+=(_wiz_setup_prompt)
fi
