package season

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/inkel/cabb"
	"github.com/inkel/cabb/cmd/cabb/messages"
)

type Model struct {
	season cabb.Season
	dates  table.Model
	games  table.Model
	board  table.Model
}

var (
	tcs  = lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)
	ncs  = lipgloss.NewStyle().AlignHorizontal(lipgloss.Right)
	bold = lipgloss.NewStyle().Bold(true)
)

func NewModel(w, h int, s cabb.Season) Model {
	m := Model{
		season: s,
		dates:  datesTable(w/2, h/2, s.Season),
		board:  boardTable(w/2, s.Positions),
		games:  gamesTable(w, h/2),
	}

	for _, gm := range s.Season {
		if gm.Current {
			m = m.withGames(gm.Matches)
		}
	}

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch {
	case m.dates.GetFocused():
		m.dates, cmd = m.dates.Update(msg)
		cmds = append(cmds, cmd)

	case m.games.GetFocused():
		m.games, cmd = m.games.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			m.dates = m.dates.Focused(!m.dates.GetFocused())
			m.games = m.games.Focused(!m.games.GetFocused())

		case tea.KeyEnter:
			if m.games.GetFocused() {
				return m, messages.Load(m.games.HighlightedRow().Data["Match"])
			}

		case tea.KeyEscape:
			return m, messages.Back

		case tea.KeyCtrlR:
			return m, messages.Load(cabb.Team{ID: m.season.TeamID})

		case tea.KeyUp, tea.KeyDown:
			if m.dates.GetFocused() {
				gd := m.dates.HighlightedRow().Data["GameDay"].(cabb.GameDay)
				m = m.withGames(gd.Matches)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var s strings.Builder

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left, m.dates.View(), m.board.View()),
		m.games.View(), s.String())
}

func datesTable(w, h int, dates []cabb.GameDay) table.Model {
	cols := []table.Column{
		table.NewColumn("Date", "Fecha", 10),
		table.NewFlexColumn("Name", "Jornada", 1),
	}

	var hl int

	rows := make([]table.Row, len(dates))
	for i, d := range dates {
		d := d
		rows[i] = table.NewRow(table.RowData{
			"GameDay": d,
			"Date":    d.Date,
			"Name":    d.Name,
		})
		if d.Current {
			hl = i
			rows[i] = rows[i].WithStyle(bold)
		}
	}

	return table.New(cols).WithRows(rows).
		WithTargetWidth(w).
		WithPageSize(h).
		WithBaseStyle(tcs).
		Focused(true).
		WithHighlightedRow(hl)
}

func boardTable(width int, pos []cabb.Position) table.Model {
	cols := []table.Column{
		table.NewFlexColumn("Name", "Nombre", 2).WithStyle(tcs),
		table.NewColumn("PJ", "PJ", 2).WithStyle(ncs),
		table.NewColumn("PG", "PG", 2).WithStyle(ncs),
		table.NewColumn("PP", "PP", 2).WithStyle(ncs),
		table.NewColumn("PF", "PF", 3).WithStyle(ncs),
		table.NewColumn("PC", "PC", 3).WithStyle(ncs),
		table.NewColumn("APF", "PPF", 3).WithStyle(ncs),
		table.NewColumn("APC", "PPC", 3).WithStyle(ncs),
		table.NewColumn("APD", "PDP", 4).WithStyle(ncs),
		table.NewColumn("PS", "Puntos", 6).WithStyle(ncs),
	}

	rows := make([]table.Row, len(pos))
	for i, p := range pos {
		rows[i] = table.NewRow(table.RowData{
			"ID":   p.ID,
			"Name": p.Name,
			"PJ":   p.Played,
			"PG":   p.Won,
			"PP":   p.Lost,
			"PF":   p.Scored,
			"PC":   p.Received,
			"APF":  p.Scored / p.Played,
			"APC":  p.Received / p.Played,
			"APD":  (p.Scored - p.Received) / p.Played,
			"PS":   p.Score,
		})
	}

	return table.New(cols).WithRows(rows).WithTargetWidth(width).Focused(false)
}

func gamesTable(w, h int) table.Model {
	cols := []table.Column{
		table.NewColumn("Date", "Fecha", 11),
		table.NewFlexColumn("Home", "Local", 2).WithStyle(tcs),
		table.NewColumn("HS", "#", 3).WithStyle(ncs),
		table.NewColumn("AS", "#", 3).WithStyle(ncs),
		table.NewFlexColumn("Away", "Visitante", 2).WithStyle(tcs),
		table.NewFlexColumn("Status", "Estado", 1).WithStyle(tcs),
	}

	return table.New(cols).
		WithPageSize(h).
		WithTargetWidth(w)
}

func (m Model) withGames(games []cabb.Match) Model {
	rows := make([]table.Row, len(games))
	for i, g := range games {
		var d, t string
		if g.Date != "" {
			d = g.Date[0:5]
		}
		if g.Time != "" {
			t = g.Time[0:5]
		}
		rows[i] = table.NewRow(table.RowData{
			"Match":  g,
			"Date":   d + " " + t,
			"Home":   g.HomeTeam,
			"HS":     g.HomeScore,
			"AS":     g.AwayScore,
			"Away":   g.AwayTeam,
			"Status": g.Status,
		})
	}
	m.games = m.games.WithRows(rows)
	return m
}

func (m Model) ShortHelp() []key.Binding { return nil }

func (m Model) FullHelp() [][]key.Binding { return nil }
