package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inkel/cabb"
	"github.com/inkel/cabb/cmd/cabb/messages"
	"github.com/inkel/cabb/cmd/cabb/pages/live"
	"github.com/inkel/cabb/cmd/cabb/pages/season"
	"github.com/inkel/cabb/cmd/cabb/pages/stats"
	"github.com/inkel/cabb/cmd/cabb/pages/teams"
)

func dieIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	cabb.D = true

	var (
		uid      = os.Getenv("CABBUID")
		deviceID = os.Getenv("DEVICEID")
	)

	c, err := cabb.NewClient(uid, deviceID)
	dieIf(err)

	m := model{
		client:  c,
		spinner: spinner.New(spinner.WithSpinner(spinner.Points)),
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		dieIf(err)
	}
}

type page int

const (
	pageLoading page = iota
	pageTeams
	pageSeason
	pageStats
	pageLive
)

type model struct {
	page   page
	msg    string
	err    error
	client cabb.Client
	w, h   int

	spinner spinner.Model
	teams   teams.Model
	season  season.Model
	stats   stats.Model
	live    live.Model
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen,
		messages.Loading("Cargando equipos"),
		m.loadTeams)
	// messages.Load(cabb.Team{ID: "38007500480071003700730055004100460050005100320062004D00660068006C004900390035002B0051003D003D00"}))
	//m.loadTeams)
	// return tea.Batch(
	// 	tea.EnterAltScreen,
	// 	m.loadMatch(cabb.Match{MatchID: "58004F00640051006C00750074004700390051007900380053005A0058006F0031006E004C006F00740077003D003D00"}),
	// )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case messages.LoadingMsg:
		m.page = pageLoading
		m.msg = string(msg)
		return m, m.spinner.Tick

	case cabb.Team:
		return m, tea.Batch(messages.Loading("Cargando temporada %s", msg.Name), m.loadSeason(msg))

	case cabb.Match:
		return m, tea.Batch(messages.Loading("Cargando partido %s - %s", msg.HomeTeam, msg.AwayTeam), m.loadMatch(msg))

	case []cabb.Team:
		m.page = pageTeams
		m.teams = teams.NewModel(m.w, m.h, msg)
		return m, nil

	case cabb.Season:
		m.page = pageSeason
		m.season = season.NewModel(m.w, m.h, msg)

	case cabb.Stats:
		m.page = pageStats
		m.stats = stats.New(m.w, m.h, msg)

	case cabb.Live:
		m.page = pageLive
		m.live = live.New(m.w, m.h, msg)

	case messages.BackMsg:
		m.page = pageSeason

	case messages.LiveMatchMsg:
		return m, m.liveMatch(msg.Match)

	case error:
		m.err = msg
	}

	switch m.page {
	case pageTeams:
		m.teams, cmd = m.teams.Update(msg)

	case pageSeason:
		m.season, cmd = m.season.Update(msg)

	case pageStats:
		m.stats, cmd = m.stats.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	switch m.page {
	case pageLoading:
		return fmt.Sprintf("%s %s\n", m.msg, m.spinner.View())

	case pageTeams:
		return m.teams.View()

	case pageSeason:
		return m.season.View()

	case pageStats:
		return m.stats.View()

	case pageLive:
		return m.live.View()
	}

	return fmt.Sprintf("NO PAGE FOUND FOR %v\n", m.page)
}

func (m model) loadTeams() tea.Msg {
	ts, err := fetch("teams.json", m.client.Teams)
	if err != nil {
		return err
	}

	return ts
}

func (m model) loadSeason(team cabb.Team) tea.Cmd {
	return func() tea.Msg {
		s, err := m.client.Season(team.ID)
		if err != nil {
			return fmt.Errorf("loading season for team %s: %w", team.Name, err)
		}

		return s
	}
}

func (m model) loadMatch(match cabb.Match) tea.Cmd {
	return func() tea.Msg {
		g, err := m.client.Stats(match)
		if err != nil {
			return err
		}
		return g
	}
}

func (m model) liveMatch(match cabb.Match) tea.Cmd {
	return func() tea.Msg {
		l, err := m.client.Live(match)
		if err != nil {
			return err
		}
		return l
	}
}
