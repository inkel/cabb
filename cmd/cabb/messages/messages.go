package messages

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inkel/cabb"
)

type LoadingMsg string

func Loading(msg string, args ...any) tea.Cmd {
	return func() tea.Msg {
		return LoadingMsg(fmt.Sprintf(msg, args...))
	}
}

func Load[T any](t T) tea.Cmd { return func() tea.Msg { return t } }

type BackMsg int

var Back tea.Cmd = Load(BackMsg(0))

type LiveMatchMsg struct {
	Match cabb.Match
}

func LiveMatch(match cabb.Match) tea.Cmd {
	return func() tea.Msg { return LiveMatchMsg{match} }
}
