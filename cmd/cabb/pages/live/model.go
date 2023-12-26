package live

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inkel/cabb"
	"github.com/inkel/cabb/cmd/cabb/messages"
)

type Model struct {
	live cabb.Live
	view viewport.Model
}

func New(w, h int, l cabb.Live) Model {
	v := viewport.New(w, h)
	v.SetContent(live(l))

	return Model{
		live: l,
		view: v,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "g" {
			return m, messages.LiveMatch(m.live.Match)
		}
		if msg.Type == tea.KeyEsc {
			return m, messages.Back
		}
	}

	m.view, cmd = m.view.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.view.View()
}

func live(l cabb.Live) string {
	var s strings.Builder

	ts := map[int]string{
		l.LiveMatch.HomeID: l.LiveMatch.Home,
		l.LiveMatch.AwayID: l.LiveMatch.Away,
	}

	homeID := l.LiveMatch.HomeID

	w := tabwriter.NewWriter(&s, 2, 2, 1, ' ', 0)

	fmt.Fprintf(w, "CUARTO\t%s\tJUGADOR\t%s\n", ts[l.LiveMatch.HomeID], ts[l.LiveMatch.AwayID])

	for _, a := range l.Live.Actions {
		if a.TeamID == homeID {
			fmt.Fprintf(w, "%d\t%s\t%s\t-\n", a.Period, a.Type, a.PlayerNum)
		} else {
			fmt.Fprintf(w, "%d\t-\t%s\t%s\n", a.Period, a.PlayerNum, a.Type)
		}
	}

	if err := w.Flush(); err != nil {
		return err.Error()
	}

	return s.String()
}
