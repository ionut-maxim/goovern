package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ionut-maxim/goovern"
	"github.com/ionut-maxim/goovern/db"
)

type viewMode int

const (
	searchMode viewMode = iota
	resultsMode
	modalMode
)

type Model struct {
	textInput        textinput.Model
	pool             *pgxpool.Pool
	dbClient         *db.DB
	results          []goovern.Company
	searching        bool
	err              error
	table            table.Model
	mode             viewMode
	selectedCompany  *goovern.Company
	titleStyle       lipgloss.Style
	borderStyle      lipgloss.Style
	inputBorderStyle lipgloss.Style
	helpStyle        lipgloss.Style
	errorStyle       lipgloss.Style
	modalStyle       lipgloss.Style
	labelStyle       lipgloss.Style
	width            int
	height           int
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

type searchResultMsg struct {
	results []goovern.Company
	err     error
}
