package viewlist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Msg []string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}

	for i := 0; i < len(m.Msg); i++ {
		s.WriteString(fmt.Sprintf("%s\n", m.Msg[i]))
	}
	s.WriteString("\n[press enter to continue]")

	return s.String()
}
