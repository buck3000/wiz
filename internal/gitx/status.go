package gitx

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// RepoStatus is the parsed result of git status --porcelain=v2 -b.
type RepoStatus struct {
	Branch    string
	OID       string
	Upstream  string
	Ahead     int
	Behind    int
	Dirty     bool
	Staged    int
	Unstaged  int
	Untracked int
	Conflicted int
}

// Status runs git status --porcelain=v2 -b and parses the output.
func (r *Repo) Status(ctx context.Context) (*RepoStatus, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain=v2", "-b")
	cmd.Dir = r.WorkDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	return ParseStatusPorcelainV2(string(out))
}

// StatusAt runs git status in the given directory.
func StatusAt(ctx context.Context, dir string) (*RepoStatus, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain=v2", "-b")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	return ParseStatusPorcelainV2(string(out))
}

// ParseStatusPorcelainV2 parses the output of git status --porcelain=v2 -b.
func ParseStatusPorcelainV2(output string) (*RepoStatus, error) {
	s := &RepoStatus{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			s.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.oid "):
			s.OID = strings.TrimPrefix(line, "# branch.oid ")
		case strings.HasPrefix(line, "# branch.upstream "):
			s.Upstream = strings.TrimPrefix(line, "# branch.upstream ")
		case strings.HasPrefix(line, "# branch.ab "):
			ab := strings.TrimPrefix(line, "# branch.ab ")
			parseAB(ab, s)
		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			parseChangedEntry(line, s)
			s.Dirty = true
		case strings.HasPrefix(line, "u "):
			s.Conflicted++
			s.Dirty = true
		case strings.HasPrefix(line, "? "):
			s.Untracked++
			s.Dirty = true
		}
	}
	return s, scanner.Err()
}

func parseAB(ab string, s *RepoStatus) {
	parts := strings.Fields(ab)
	if len(parts) >= 2 {
		s.Ahead, _ = strconv.Atoi(strings.TrimPrefix(parts[0], "+"))
		behind, _ := strconv.Atoi(strings.TrimPrefix(parts[1], "-"))
		s.Behind = behind
	}
}

func parseChangedEntry(line string, s *RepoStatus) {
	// Format: 1 XY ... or 2 XY ...
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	xy := fields[1]
	if len(xy) < 2 {
		return
	}
	x := xy[0]
	y := xy[1]
	if x != '.' {
		s.Staged++
	}
	if y != '.' {
		s.Unstaged++
	}
}
