package app

import (
	"context"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ionut-maxim/goovern/db"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case searchMode:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyEnter:
				searchTerm := m.textInput.Value()
				if searchTerm != "" {
					m.searching = true
					m.err = nil
					return m, performSearch(m.pool, m.dbClient, searchTerm)
				}
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case resultsMode:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.mode = searchMode
				m.textInput.Focus()
				return m, nil
			case tea.KeyEnter:
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 && len(m.results) > m.table.Cursor() {
					m.selectedCompany = &m.results[m.table.Cursor()]
					m.mode = modalMode
					return m, nil
				}
			}
			m.table, cmd = m.table.Update(msg)
			return m, cmd

		case modalMode:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc, tea.KeyEnter:
				m.mode = resultsMode
				m.selectedCompany = nil
				return m, nil
			}
		}

	case searchResultMsg:
		m.searching = false
		m.results = msg.results
		m.err = msg.err

		if msg.err == nil && len(msg.results) > 0 {
			rows := make([]table.Row, len(msg.results))
			for i, comp := range msg.results {
				rows[i] = table.Row{
					truncate(comp.Name, 35),
					comp.TaxID,
					comp.RegistrationCode,
					truncate(comp.LegalForm, 20),
				}
			}
			m.table.SetRows(rows)
			m.mode = resultsMode
			m.textInput.Blur()
		}
		return m, nil
	}

	return m, cmd
}

func performSearch(pool *pgxpool.Pool, dbClient *db.DB, searchTerm string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		results, err := dbClient.Search(ctx, pool, searchTerm, 10)
		return searchResultMsg{
			results: results,
			err:     err,
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
