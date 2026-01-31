# wiz shell integration for fish
# Add to config.fish: wiz init fish | source

function wiz
    if test "$argv[1]" = "enter"
        set -l output (command wiz enter $argv[2..])
        or return $status
        eval $output
    else
        command wiz $argv
    end
end

# Cache variables for status optimization.
set -g _wiz_last_index_mtime ""
set -g _wiz_cached_status ""

function _wiz_prompt_hook --on-event fish_prompt
    if test -n "$WIZ_CTX"
        # Only call wiz status if the git index has changed.
        set -l index_file "$WIZ_DIR/.git/index"
        if not test -f "$index_file"
            set index_file "$WIZ_DIR/../.git/worktrees/$WIZ_CTX/index"
        end
        set -l current_mtime ""
        if test -f "$index_file"
            set current_mtime (command stat -f %m "$index_file" 2>/dev/null; or command stat -c %Y "$index_file" 2>/dev/null)
        end

        if test "$current_mtime" != "$_wiz_last_index_mtime"; or test -z "$_wiz_cached_status"
            set -g _wiz_cached_status (command wiz status --porcelain 2>/dev/null)
            set -g _wiz_last_index_mtime "$current_mtime"
        end

        if test -n "$_wiz_cached_status"
            set -l parts (string split ' ' $_wiz_cached_status)
            set -l ctx $parts[1]
            set -l repo $parts[2]
            set -l branch $parts[3]
            set -l state $parts[4]
            set -l dirty ""
            if test "$state" = "dirty"
                set dirty "*"
            end
            # Terminal title
            printf '\033]0;🧙 %s — %s\007' $ctx $repo
            set -gx WIZ_PROMPT "🧙 $ctx$dirty"
        end
    end
end

# Save original prompt
if functions -q fish_prompt; and not functions -q _wiz_original_fish_prompt
    functions -c fish_prompt _wiz_original_fish_prompt
end

function fish_prompt
    if test -n "$WIZ_PROMPT"
        set_color magenta
        echo -n "$WIZ_PROMPT "
        set_color normal
    end
    # Call the original prompt if it exists
    if functions -q _wiz_original_fish_prompt
        _wiz_original_fish_prompt
    else
        echo -n '> '
    end
end
