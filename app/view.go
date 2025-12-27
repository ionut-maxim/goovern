package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.mode == modalMode && m.selectedCompany != nil {
		return m.renderModal()
	}

	var content strings.Builder

	switch m.mode {
	case searchMode:
		content.WriteString(m.renderSearchView())
	case resultsMode:
		content.WriteString(m.renderResultsView())
	default:
		panic("unhandled default case")
	}

	return content.String()
}

func (m Model) renderSearchView() string {
	var content strings.Builder

	// ASCII art logo in lighter purple
	logo := `
  ██████╗  ██████╗  ██████╗ ██╗   ██╗███████╗██████╗ ███╗   ██╗
 ██╔════╝ ██╔═══██╗██╔═══██╗██║   ██║██╔════╝██╔══██╗████╗  ██║
 ██║  ███╗██║   ██║██║   ██║██║   ██║█████╗  ██████╔╝██╔██╗ ██║
 ██║   ██║██║   ██║██║   ██║╚██╗ ██╔╝██╔══╝  ██╔══██╗██║╚██╗██║
 ╚██████╔╝╚██████╔╝╚██████╔╝ ╚████╔╝ ███████╗██║  ██║██║ ╚████║
  ╚═════╝  ╚═════╝  ╚═════╝   ╚═══╝  ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝`

	logoStyle := m.titleStyle.UnsetBackground().Foreground(lipgloss.Color("#9D7CD8"))
	content.WriteString("\n")
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logoStyle.Render(logo)))
	content.WriteString("\n\n")

	subtitle := m.helpStyle.Render("Romanian Company Search")
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, subtitle))
	content.WriteString("\n\n")

	inputBox := m.inputBorderStyle.Render(m.textInput.View())
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, inputBox))
	content.WriteString("\n\n")

	var statusMsg string
	if m.searching {
		statusMsg = "Searching..."
	} else if m.err != nil {
		statusMsg = m.errorStyle.Render("Error: " + m.err.Error())
	} else if len(m.results) > 0 {
		statusMsg = fmt.Sprintf("Found %d results. Press Enter to view.", len(m.results))
	}
	if statusMsg != "" {
		content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, statusMsg))
		content.WriteString("\n")
	}

	help := m.helpStyle.Render("enter: search • ctrl+c: quit")
	content.WriteString("\n")
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help))

	return content.String()
}

func (m Model) renderResultsView() string {
	var content strings.Builder

	title := m.titleStyle.Render("  Search Results  ")
	content.WriteString("\n")
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, title))
	content.WriteString("\n\n")

	tableView := m.borderStyle.Render(m.table.View())
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, tableView))
	content.WriteString("\n")

	help := m.helpStyle.Render("↑/↓: navigate • enter: view details • esc: back to search • ctrl+c: quit")
	content.WriteString("\n")
	content.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help))

	return content.String()
}

func (m Model) renderModal() string {
	if m.selectedCompany == nil {
		return ""
	}

	c := m.selectedCompany
	var content strings.Builder

	title := m.titleStyle.Render("  Company Details  ")
	content.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, title))
	content.WriteString("\n\n")

	renderField := func(label, value string) {
		if value != "" {
			content.WriteString(m.labelStyle.Render(label) + "\n")
			content.WriteString("  " + value + "\n\n")
		}
	}

	renderField("Company Name", c.Name)
	renderField("Tax ID (CUI)", c.TaxID)
	renderField("Registration Code", c.RegistrationCode)
	renderField("Registration Date", c.RegistrationDate)
	renderField("Legal Form", c.LegalForm)
	renderField("EUID", c.EUID)

	if c.Country != "" || c.County != "" || c.Locality != "" {
		content.WriteString(m.labelStyle.Render("Location") + "\n")
		var location []string
		if c.Locality != "" {
			location = append(location, c.Locality)
		}
		if c.County != "" {
			location = append(location, c.County)
		}
		if c.Country != "" {
			location = append(location, c.Country)
		}
		if c.Sector != "" {
			location = append(location, "Sector "+c.Sector)
		}
		content.WriteString("  " + strings.Join(location, ", ") + "\n\n")
	}

	if c.StreetName != "" || c.StreetNumber != "" || c.Building != "" ||
		c.Staircase != "" || c.Floor != "" || c.Apartment != "" || c.PostalCode != "" {
		content.WriteString(m.labelStyle.Render("Address") + "\n")
		var address []string
		if c.StreetName != "" {
			streetAddr := c.StreetName
			if c.StreetNumber != "" {
				streetAddr += " " + c.StreetNumber
			}
			address = append(address, streetAddr)
		}
		if c.Building != "" {
			address = append(address, "Building "+c.Building)
		}
		if c.Staircase != "" {
			address = append(address, "Staircase "+c.Staircase)
		}
		if c.Floor != "" {
			address = append(address, "Floor "+c.Floor)
		}
		if c.Apartment != "" {
			address = append(address, "Apt "+c.Apartment)
		}
		if c.PostalCode != "" {
			address = append(address, "Postal Code: "+c.PostalCode)
		}
		content.WriteString("  " + strings.Join(address, ", ") + "\n\n")
	}

	renderField("Address Details", c.AddressDetails)
	renderField("Website", c.Website)
	renderField("Parent Company Country", c.ParentCompanyCountry)

	content.WriteString(m.labelStyle.Render("Search Rank Score") + "\n")
	content.WriteString(fmt.Sprintf("  %.6f\n", c.Rank))

	modalContent := m.modalStyle.Render(content.String())

	help := m.helpStyle.Render("enter/esc: close • ctrl+c: quit")

	var fullView strings.Builder
	fullView.WriteString("\n\n")
	fullView.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, modalContent))
	fullView.WriteString("\n\n")
	fullView.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help))

	return fullView.String()
}
