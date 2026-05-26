package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/jwt-tui/internal/keys"
	"github.com/anotherhadi/jwt-tui/internal/util"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.keymap.width = msg.Width
		m.recalcSizes()
		if m.showDocs {
			m.renderDocs()
		}
		return m, nil

	case tea.ClipboardMsg:
		content := msg.String()
		if content != "" {
			panel := m.pendingPastePanel
			m.panels[panel].ta.SetValue(content)
			m.setViewportContent(panel, content)
			switch panel {
			case panelHeader, panelPayload:
				m.rebuildJWT()
			case panelJWT:
				m.decodeJWT()
			case panelSecret:
				m.revalidate()
			}
		}
		return m, nil

	case util.EditorFinishedMsg:
		if msg.Err == nil && msg.Content != "" {
			panel := m.pendingEditorPanel
			m.panels[panel].ta.SetValue(msg.Content)
			m.setViewportContent(panel, msg.Content)
			switch panel {
			case panelHeader, panelPayload:
				m.rebuildJWT()
			case panelJWT:
				m.decodeJWT()
			case panelSecret:
				m.revalidate()
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		// Docs overlay: only scrolling, d or esc to close, ctrl+c to quit
		if m.showDocs {
			switch {
			case key.Matches(msg, keys.Keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, keys.Keys.Docs), msg.String() == "esc":
				m.showDocs = false
			default:
				var cmd tea.Cmd
				m.docsVP, cmd = m.docsVP.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// In edit mode: esc and ctrl+c exit edit mode, everything else goes to the textarea
		if m.panels[m.focus].editing {
			if msg.String() == "esc" || msg.String() == "ctrl+c" {
				m.exitEditMode()
				return m, nil
			}
			p := &m.panels[m.focus]
			prev := p.ta.Value()
			var cmd tea.Cmd
			p.ta, cmd = p.ta.Update(msg)
			if p.ta.Value() != prev {
				switch m.focus {
				case panelHeader, panelPayload:
					m.rebuildJWT()
				case panelJWT:
					m.decodeJWT()
				case panelSecret:
					m.revalidate()
				}
			}
			return m, cmd
		}

		// View mode shortcuts
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Keys.HelpToggle):
			m.help.ShowAll = !m.help.ShowAll
			m.recalcSizes()

		case key.Matches(msg, keys.Keys.Docs):
			m.help.ShowAll = false
			m.showDocs = true
			m.renderDocs()

		case key.Matches(msg, keys.Keys.CycleFocus):
			m.focus = (m.focus + 1) % 4

		case key.Matches(msg, keys.Keys.Edit):
			return m, m.enterEditMode()

		case key.Matches(msg, keys.Keys.EditExternal):
			m.pendingEditorPanel = m.focus
			return m, util.OpenExternalEditor(m.panels[m.focus].ta.Value())

		case key.Matches(msg, keys.Keys.Copy):
			return m, tea.SetClipboard(m.panels[m.focus].ta.Value())

		case key.Matches(msg, keys.Keys.Paste):
			m.pendingPastePanel = m.focus
			return m, tea.ReadClipboard

		case key.Matches(msg, keys.Keys.Clear):
			return m, m.clearPanel()

		case key.Matches(msg, keys.Keys.Reset):
			m.resetPanel()

		default:
			var cmd tea.Cmd
			m.panels[m.focus].vp, cmd = m.panels[m.focus].vp.Update(msg)
			return m, cmd
		}
		return m, nil

	default:
		if m.panels[m.focus].editing {
			p := &m.panels[m.focus]
			var cmd tea.Cmd
			p.ta, cmd = p.ta.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.panels[m.focus].vp, cmd = m.panels[m.focus].vp.Update(msg)
		return m, cmd
	}
}
