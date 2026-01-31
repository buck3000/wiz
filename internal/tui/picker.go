package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
)

// Action is what the user chose to do in the picker.
type Action int

const (
	ActionNone Action = iota
	ActionEnter
	ActionSpawn
	ActionDelete
	ActionCreate
)

// Result holds the picker outcome.
type Result struct {
	Action  Action
	Context *wizctx.Context
}

// Model is the Bubble Tea model for the context picker.
type Model struct {
	contexts []wizctx.Context
	filtered []int // indices into contexts
	cursor   int
	result   Result
	quitting bool
	width    int
	height   int
	filter   string
	filtering bool
}

// NewPicker creates a new picker model.
func NewPicker(contexts []wizctx.Context) Model {
	indices := make([]int, len(contexts))
	for i := range contexts {
		indices[i] = i
	}
	return Model{
		contexts: contexts,
		filtered: indices,
	}
}

func (m *Model) applyFilter() {
	if m.filter == "" {
		m.filtered = make([]int, len(m.contexts))
		for i := range m.contexts {
			m.filtered[i] = i
		}
	} else {
		m.filtered = m.filtered[:0]
		lower := strings.ToLower(m.filter)
		for i, c := range m.contexts {
			if strings.Contains(strings.ToLower(c.Name), lower) ||
				strings.Contains(strings.ToLower(c.Branch), lower) ||
				strings.Contains(strings.ToLower(c.Task), lower) ||
				strings.Contains(strings.ToLower(c.Agent), lower) {
				m.filtered = append(m.filtered, i)
			}
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter":
				m.filtering = false
			case "esc":
				m.filter = ""
				m.filtering = false
				m.applyFilter()
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.filter += msg.String()
					m.applyFilter()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "/":
			m.filtering = true

		case "enter":
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.result = Result{Action: ActionEnter, Context: &m.contexts[idx]}
				m.quitting = true
				return m, tea.Quit
			}

		case "s":
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.result = Result{Action: ActionSpawn, Context: &m.contexts[idx]}
				m.quitting = true
				return m, tea.Quit
			}

		case "d":
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.result = Result{Action: ActionDelete, Context: &m.contexts[idx]}
				m.quitting = true
				return m, tea.Quit
			}

		case "n":
			m.result = Result{Action: ActionCreate}
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("\U0001f9d9 wiz contexts"))
	b.WriteString("\n")

	if m.filtering || m.filter != "" {
		filterDisplay := m.filter
		if m.filtering {
			filterDisplay += "\u2588" // cursor block
		}
		b.WriteString(dimStyle.Render(fmt.Sprintf("filter: %s", filterDisplay)))
		b.WriteString("\n")
	}

	if len(m.filtered) == 0 {
		if m.filter != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf("No matches for %q. Press esc to clear.", m.filter)))
		} else {
			b.WriteString(dimStyle.Render("No contexts. Press 'n' to create one."))
		}
		b.WriteString("\n")
	} else {
		curCtx := os.Getenv("WIZ_CTX")
		for vi, ci := range m.filtered {
			c := m.contexts[ci]
			cursor := "  "
			if vi == m.cursor {
				cursor = "\u25b8 "
			}
			active := ""
			if c.Name == curCtx {
				active = " \u2022"
			}

			line := fmt.Sprintf("%s%s%s", cursor, c.Name, active)
			if vi == m.cursor {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(normalStyle.Render(line))
			}
			b.WriteString("\n")

			// Detail line: branch, strategy, and optionally agent/task.
			detail := fmt.Sprintf("branch: %s | %s", c.Branch, c.Strategy)
			if c.Agent != "" {
				detail += fmt.Sprintf(" | agent: %s", c.Agent)
			}
			b.WriteString(dimStyle.Render(detail))
			b.WriteString("\n")

			if c.Task != "" && vi == m.cursor {
				taskLine := c.Task
				if len(taskLine) > 60 {
					taskLine = taskLine[:57] + "..."
				}
				b.WriteString(dimStyle.Render(fmt.Sprintf("task: %s", taskLine)))
				b.WriteString("\n")
			}
		}
	}

	help := "enter: activate  s: spawn  d: delete  n: new  /: filter  q: quit"
	b.WriteString(helpStyle.Render(help))
	b.WriteString("\n")

	return b.String()
}

// Result returns the picker result. Call after tea.Program finishes.
func (m Model) Result() Result {
	return m.result
}

// Run launches the TUI picker and returns the result.
func Run(contexts []wizctx.Context) (Result, error) {
	m := NewPicker(contexts)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return Result{}, err
	}
	return finalModel.(Model).Result(), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
