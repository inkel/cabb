package stats

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/inkel/cabb"
	"github.com/inkel/cabb/cmd/cabb/messages"
)

type Model struct {
	stats cabb.Stats
	score table.Model
	home  table.Model
	away  table.Model
}

func New(w, h int, stats cabb.Stats) Model {
	score := matchScore(stats)

	w = w - lipgloss.Width(score.View())

	return Model{
		stats: stats,
		score: score,
		home:  playerStats(w, stats.Stats.Home).Focused(true),
		away:  playerStats(w, stats.Stats.Away),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEscape {
			cmd = messages.Back
			break
		} else if msg.Type == tea.KeyTab {
			m.home = m.home.Focused(!m.home.GetFocused())
			m.away = m.away.Focused(!m.away.GetFocused())
			break
		} else if msg.Type == tea.KeyCtrlR {
			return m, messages.Load(cabb.Match{MatchID: m.stats.MatchID})
		} else if msg.Type == tea.KeyCtrlL {
			return m, messages.LiveMatch(cabb.Match{MatchID: m.stats.MatchID})
		}

		if m.home.GetFocused() {
			m.home, cmd = m.home.Update(msg)
		} else {
			m.away, cmd = m.away.Update(msg)
		}
	}

	return m, cmd
}

func (m Model) View() string {
	players := lipgloss.JoinVertical(lipgloss.Top, m.home.View(), m.away.View())
	data := lipgloss.JoinHorizontal(lipgloss.Left, m.score.View(), players)
	return lipgloss.JoinVertical(lipgloss.Top,
		m.stats.MatchID,
		fmt.Sprintf("%s %3d - %3d %s", m.stats.Match.Home, m.stats.Match.HomeScore, m.stats.Match.AwayScore, m.stats.Match.Away),
		data)
}

var (
	tcs = lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)
	ncs = lipgloss.NewStyle().AlignHorizontal(lipgloss.Right)

	totals = lipgloss.NewStyle().Bold(true)
)

func playerStats(w int, players []cabb.PlayerStats) table.Model {
	nc := func(id, hdr string, w int) table.Column {
		return table.NewColumn(id, hdr, w).WithStyle(ncs)
	}

	cols := []table.Column{
		nc("No", "#", 2),
		table.NewFlexColumn("Name", "Nombre", 1).WithStyle(tcs),
		nc("Played", "Mins", 5),
		nc("Points", "PS", 3),
		nc("2P", "2P", 12),
		nc("1P", "1P", 12),
		nc("3P", "3P", 12),
		nc("A", "AS", 2),
		nc("TO", "TO", 2),
		nc("F", "F", 2),
		nc("R", "R", 2),
		nc("RB", "RT", 2),
		nc("RBO", "RO", 2),
		nc("RBD", "RD", 2),
		nc("VAL", "VAL", 3),
	}

	shots := func(m, a int) string {
		if a == 0 {
			return ""
		}
		p := float64(m) / float64(a)
		return fmt.Sprintf("%2d/%2d (%.2f)", m, a, p)
	}

	height := len(players)
	rows := make([]table.Row, height)
	for i, p := range players {
		rows[i] = table.NewRow(table.RowData{
			"No":     p.Num,
			"Name":   p.Name,
			"Played": p.Played,
			"Points": p.Points,
			"RB":     p.Rebounds,
			"RBO":    p.ReboundsOff,
			"RBD":    p.ReboundsDef,
			"F":      p.Fouls,
			"R":      p.Fouled,
			"A":      p.Assists,
			"TO":     p.Turnovers,
			"2P":     shots(p.Made2P, p.Shots2P),
			"1P":     shots(p.Made1P, p.Shots1P),
			"3P":     shots(p.Made3P, p.Shots3P),
			"VAL":    p.Val,
		})
		if p.Num == "" && p.Name == "TOTALES" {
			rows[i] = rows[i].WithStyle(totals)
			delete(rows[i].Data, "Played")
		}
	}

	return table.New(cols).WithRows(rows).
		WithTargetWidth(w).
		WithPageSize(height).
		WithFooterVisibility(false)
}

func matchScore(s cabb.Stats) table.Model {
	cols := []table.Column{
		table.NewColumn("P", "#", 2),
		table.NewColumn("H", "L", 3),
		table.NewColumn("A", "V", 3),
		table.NewColumn("TH", "TL", 2),
		table.NewColumn("TA", "TV", 2),
		table.NewColumn("PD", "DP", 3),
		table.NewColumn("TD", "DT", 3),
	}

	rows := make([]table.Row, len(s.Match.Periods))

	var h, a int

	for i, p := range s.Match.Periods {
		h += p.HomeScore
		a += p.AwayScore

		rows[i] = table.NewRow(table.RowData{
			"P":  p.Period,
			"H":  p.HomeScore,
			"A":  p.AwayScore,
			"TH": h,
			"TA": a,
			"PD": p.HomeScore - p.AwayScore,
			"TD": h - a,
		})
	}

	return table.New(cols).WithRows(rows).
		WithPageSize(len(s.Match.Periods)).
		WithBaseStyle(ncs).
		WithFooterVisibility(false)
}
