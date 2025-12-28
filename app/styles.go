package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muesli/termenv"

	"github.com/ionut-maxim/goovern/db"
)

func NewModel(s ssh.Session, pool *pgxpool.Pool, dbClient *db.DB) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	renderer := bubbletea.MakeRenderer(s)
	lipgloss.SetColorProfile(termenv.TrueColor)

	// Create styles with bubbles components aesthetic
	titleStyle := renderer.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		MarginBottom(1)

	borderStyle := renderer.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	inputBorderStyle := renderer.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	helpStyle := renderer.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		MarginTop(1)

	errorStyle := renderer.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	modalStyle := renderer.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		Width(60)

	labelStyle := renderer.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	ti := textinput.New()
	ti.Placeholder = "Search by company name or CUI..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	// Create table with bubbles styling
	columns := []table.Column{
		{Title: "Name", Width: 35},
		{Title: "CUI", Width: 12},
		{Title: "Reg Code", Width: 12},
		{Title: "Legal Form", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))
	tableStyles.Selected = tableStyles.Selected.
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Underline(true)
	t.SetStyles(tableStyles)

	m := Model{
		textInput:        ti,
		pool:             pool,
		dbClient:         dbClient,
		table:            t,
		mode:             searchMode,
		titleStyle:       titleStyle,
		borderStyle:      borderStyle,
		inputBorderStyle: inputBorderStyle,
		helpStyle:        helpStyle,
		errorStyle:       errorStyle,
		modalStyle:       modalStyle,
		labelStyle:       labelStyle,
		width:            pty.Window.Width,
		height:           pty.Window.Height,
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
