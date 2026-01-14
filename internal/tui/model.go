package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Item represents a selectable item
type Item struct {
	Name        string
	Path        string
	Description string
	IsCurrent   bool
}

// Model is the Bubbletea model for selection
type Model struct {
	title     string
	items     []Item
	filtered  []Item
	cursor    int
	textInput textinput.Model
	selected  *Item
	quitting  bool
	filtering bool
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(lipgloss.Color("170")).
			Bold(true)

	currentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	filterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))
)

// NewModel creates a new TUI model
func NewModel(title string, items []Item) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 30

	return Model{
		title:     title,
		items:     items,
		filtered:  items,
		cursor:    0,
		textInput: ti,
		filtering: false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.filtering {
				m.filtering = false
				m.textInput.Reset()
				m.filtered = m.items
				m.cursor = 0
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case "q":
			if !m.filtering {
				m.quitting = true
				return m, tea.Quit
			}

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.selected = &m.filtered[m.cursor]
			}
			return m, tea.Quit

		case "up", "k":
			if !m.filtering {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if !m.filtering {
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			}

		case "/":
			if !m.filtering {
				m.filtering = true
				m.textInput.Focus()
				return m, textinput.Blink
			}
		}
	}

	// Handle text input for filtering
	if m.filtering {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		// Apply filter
		query := strings.ToLower(m.textInput.Value())
		if query == "" {
			m.filtered = m.items
		} else {
			m.filtered = fuzzyFilter(m.items, query)
		}

		// Reset cursor if out of bounds
		if m.cursor >= len(m.filtered) {
			m.cursor = max(0, len(m.filtered)-1)
		}

		return m, cmd
	}

	return m, nil
}

// View renders the model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n")

	// Filter input
	if m.filtering {
		b.WriteString(filterStyle.Render("Filter: "))
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
	}

	// Items
	for i, item := range m.filtered {
		cursor := "  "
		style := itemStyle

		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		line := cursor + style.Render(item.Name)

		if item.Description != "" {
			line += descStyle.Render(" (" + item.Description + ")")
		}

		if item.IsCurrent {
			line += currentStyle.Render(" *")
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString(descStyle.Render("  No matches found"))
		b.WriteString("\n")
	}

	// Help
	help := "↑/k up • ↓/j down • / filter • enter select • esc/q quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// Selected returns the selected item
func (m Model) Selected() *Item {
	return m.selected
}

// fuzzyFilter filters items by query
func fuzzyFilter(items []Item, query string) []Item {
	var result []Item
	for _, item := range items {
		name := strings.ToLower(item.Name)
		if strings.Contains(name, query) || fuzzyMatch(name, query) {
			result = append(result, item)
		}
	}
	return result
}

// fuzzyMatch performs fuzzy matching
func fuzzyMatch(text, pattern string) bool {
	patternIdx := 0
	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			patternIdx++
		}
	}
	return patternIdx == len(pattern)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SelectBranch opens a TUI to select a branch
func SelectBranch(items []Item) (*Item, error) {
	m := NewModel("Select a branch:", items)
	p := tea.NewProgram(m, tea.WithOutput(nil))

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	return finalModel.(Model).Selected(), nil
}

// SelectWorktree opens a TUI to select a worktree
func SelectWorktree(items []Item) (*Item, error) {
	m := NewModel("Select a worktree:", items)
	p := tea.NewProgram(m, tea.WithOutput(nil))

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	return finalModel.(Model).Selected(), nil
}
