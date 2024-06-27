package viewauth

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var choices = []string{"Login", "Register"}

type Model struct {
	cursor int
	Choice string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			m.Choice = choices[m.cursor]
			return m, tea.Quit
		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}

	}
	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString("\n(press ctrl+c, esc or q to quit)\n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("(*) ")
		} else {
			s.WriteString("( )")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}
	return s.String()
}
