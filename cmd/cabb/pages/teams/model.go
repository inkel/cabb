package teams

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inkel/cabb"
	"github.com/inkel/cabb/cmd/cabb/messages"
)

type Model struct {
	list list.Model
}

var tableDefaultStyle = lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)

func NewModel(w, h int, teams []cabb.Team) Model {
	items := make([]list.Item, len(teams))
	for i, t := range teams {
		items[i] = item{t}
	}

	l := list.New(items, list.NewDefaultDelegate(), w, h)
	l.Title = "Mis Equipos"
	l.SetStatusBarItemName("equipo", "equipos")
	l.DisableQuitKeybindings()

	return Model{
		list: l,
	}
}

func (m Model) View() string {
	return m.list.View()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			i := m.list.SelectedItem().(item)
			return m, messages.Load(i.team)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type item struct {
	team cabb.Team
}

func (i item) Title() string       { return i.team.Name }
func (i item) Description() string { return i.team.Club }
func (i item) FilterValue() string { return i.team.Name }
