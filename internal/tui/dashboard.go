package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
)

// ContextStatus holds a context plus its live git status.
type ContextStatus struct {
	Context    wizctx.Context
	Status     *gitx.RepoStatus
	Error      error
	IndexMtime int64
}

type tickMsg time.Time
type statusMsg []ContextStatus

// DashboardModel is the Bubble Tea model for the watch dashboard.
type DashboardModel struct {
	store    *wizctx.Store
	statuses []ContextStatus
	interval time.Duration
	width    int
	height   int
	quitting bool
}

// NewDashboard creates a new dashboard model.
func NewDashboard(store *wizctx.Store, interval time.Duration) DashboardModel {
	return DashboardModel{
		store:    store,
		interval: interval,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchStatuses(),
		m.tick(),
	)
}

func (m DashboardModel) tick() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// getIndexMtime returns the mtime of the git index file for a context path.
func getIndexMtime(ctxPath string) int64 {
	// Try direct .git/index first (clone contexts).
	index := filepath.Join(ctxPath, ".git", "index")
	if info, err := os.Stat(index); err == nil {
		return info.ModTime().UnixNano()
	}
	// Try worktree index: look for a .git file pointing to the worktree dir.
	gitFile := filepath.Join(ctxPath, ".git")
	data, err := os.ReadFile(gitFile)
	if err == nil {
		line := strings.TrimSpace(string(data))
		if strings.HasPrefix(line, "gitdir: ") {
			wtDir := strings.TrimPrefix(line, "gitdir: ")
			index = filepath.Join(wtDir, "index")
			if info, err := os.Stat(index); err == nil {
				return info.ModTime().UnixNano()
			}
		}
	}
	return 0
}

func (m DashboardModel) fetchStatuses() tea.Cmd {
	store := m.store
	prev := m.statuses
	return func() tea.Msg {
		contexts, err := store.List()
		if err != nil {
			return statusMsg(nil)
		}

		// Build lookup of previous statuses by name for mtime comparison.
		prevMap := make(map[string]ContextStatus, len(prev))
		for _, s := range prev {
			prevMap[s.Context.Name] = s
		}

		statuses := make([]ContextStatus, len(contexts))
		var wg sync.WaitGroup
		for i, c := range contexts {
			wg.Add(1)
			go func(idx int, ctx wizctx.Context) {
				defer wg.Done()
				mtime := getIndexMtime(ctx.Path)

				// Skip git status if index hasn't changed since last fetch.
				if prev, ok := prevMap[ctx.Name]; ok && prev.IndexMtime == mtime && mtime != 0 {
					statuses[idx] = ContextStatus{
						Context:    ctx,
						Status:     prev.Status,
						Error:      prev.Error,
						IndexMtime: mtime,
					}
					return
				}

				st, err := gitx.StatusAt(context.Background(), ctx.Path)
				statuses[idx] = ContextStatus{
					Context:    ctx,
					Status:     st,
					Error:      err,
					IndexMtime: mtime,
				}
			}(i, c)
		}
		wg.Wait()

		return statusMsg(statuses)
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case tickMsg:
		return m, tea.Batch(m.fetchStatuses(), m.tick())
	case statusMsg:
		m.statuses = []ContextStatus(msg)
	}
	return m, nil
}

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252"))

	cellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")) // orange

	cleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")) // green

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // red
)

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("\U0001f9d9 wiz watch"))
	b.WriteString("\n\n")

	if len(m.statuses) == 0 {
		b.WriteString(dimStyle.Render("No contexts. Create one with: wiz create <name>"))
		b.WriteString("\n")
	} else {
		// Header row.
		b.WriteString(fmt.Sprintf("  %-20s %-10s %-20s %-10s %s\n",
			headerStyle.Render("CONTEXT"),
			headerStyle.Render("AGENT"),
			headerStyle.Render("BRANCH"),
			headerStyle.Render("STATE"),
			headerStyle.Render("CHANGES"),
		))
		b.WriteString(fmt.Sprintf("  %s\n", dimStyle.Render(strings.Repeat("\u2500", 76))))

		for _, cs := range m.statuses {
			var stateStr, diffStr string

			if cs.Error != nil {
				stateStr = errorStyle.Render("error")
				diffStr = ""
			} else if cs.Status != nil {
				if cs.Status.Dirty {
					stateStr = dirtyStyle.Render("dirty")
				} else {
					stateStr = cleanStyle.Render("clean")
				}
				var parts []string
				if cs.Status.Staged > 0 {
					parts = append(parts, fmt.Sprintf("+%d staged", cs.Status.Staged))
				}
				if cs.Status.Unstaged > 0 {
					parts = append(parts, fmt.Sprintf("~%d modified", cs.Status.Unstaged))
				}
				if cs.Status.Untracked > 0 {
					parts = append(parts, fmt.Sprintf("?%d untracked", cs.Status.Untracked))
				}
				if len(parts) > 0 {
					diffStr = strings.Join(parts, " ")
				} else {
					diffStr = "-"
				}
			} else {
				stateStr = dimStyle.Render("unknown")
				diffStr = ""
			}

			agentLabel := cs.Context.Agent
			if agentLabel == "" {
				agentLabel = "-"
			}

			name := cs.Context.Name
			if len(name) > 20 {
				name = name[:17] + "..."
			}
			branch := cs.Context.Branch
			if len(branch) > 20 {
				branch = branch[:17] + "..."
			}

			b.WriteString(fmt.Sprintf("  %-20s %-10s %-20s %-10s %s\n",
				cellStyle.Render(name),
				cellStyle.Render(agentLabel),
				cellStyle.Render(branch),
				stateStr,
				dimStyle.Render(diffStr),
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q: quit | refreshing every " + m.interval.String()))
	b.WriteString("\n")

	return b.String()
}

// RunDashboard launches the dashboard TUI.
func RunDashboard(store *wizctx.Store, interval time.Duration) error {
	m := NewDashboard(store, interval)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
