package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

func (m Model) View() tea.View {
	var content string
	if m.width == 0 {
		content = ""
	} else if m.showDocs {
		content = m.renderDocsView()
	} else {
		content = m.renderMainView()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) renderMainView() string {
	leftW := m.width / 2
	rightW := m.width - leftW
	helpH := m.helpHeight()
	availH := m.height - helpH - 1
	topH := availH / 2
	bottomH := availH - topH

	// Layout (clockwise from top-left):
	//   top-left=JWT   top-right=Header
	//   bot-left=Secret  bot-right=Payload
	jwtPanel := m.renderPanel(panelJWT, m.panelTitle(panelJWT, "Encoded"), leftW, topH)
	headerPanel := m.renderPanel(panelHeader, m.panelTitle(panelHeader, "Header"), rightW, topH)
	payloadPanel := m.renderPanel(panelPayload, m.panelTitle(panelPayload, "Payload"), rightW, bottomH)
	secretPanel := m.renderPanel(panelSecret, m.secretTitle(), leftW, bottomH)

	left := lipgloss.JoinVertical(lipgloss.Left, jwtPanel, secretPanel)
	right := lipgloss.JoinVertical(lipgloss.Left, headerPanel, payloadPanel)
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return lipgloss.JoinVertical(lipgloss.Left, main, m.renderErrorLine(), m.renderHelpBar())
}

func (m Model) renderDocsView() string {
	docsBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ilovetui.S.Subtle).
		Padding(0, 1)

	window := docsBorder.Render(ilovetui.ViewportView(&m.docsVP))
	helpStr := m.help.View(m.docsKeys)
	return lipgloss.JoinVertical(lipgloss.Left, window, helpStr)
}

func (m Model) renderErrorLine() string {
	if m.errMsg == "" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(ilovetui.S.Error).Render(" " + m.errMsg)
}

func (m Model) renderHelpBar() string {
	helpStr := m.help.View(m.keymap)

	var sigStr string
	if m.sigValid == nil {
		sigStr = ilovetui.S.Faint.Render("-")
	} else if *m.sigValid {
		sigStr = lipgloss.NewStyle().Foreground(ilovetui.S.Success).Bold(true).Render(m.sigStatus)
	} else {
		sigStr = lipgloss.NewStyle().Foreground(ilovetui.S.Error).Bold(true).Render(m.sigStatus)
	}

	// Align sig status to the right of the last line of helpStr
	helpLines := strings.Split(helpStr, "\n")
	lastLine := helpLines[len(helpLines)-1]
	lastLineW := lipgloss.Width(lastLine)
	sigW := lipgloss.Width(sigStr)
	pad := m.width - lastLineW - sigW
	if pad < 1 {
		pad = 1
	}
	helpLines[len(helpLines)-1] = lastLine + strings.Repeat(" ", pad) + sigStr
	return strings.Join(helpLines, "\n")
}
