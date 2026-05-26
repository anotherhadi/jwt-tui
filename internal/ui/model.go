package ui

import (
	_ "embed"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/jwt-tui/internal/highlight"
	"github.com/anotherhadi/jwt-tui/internal/jwt"
	"github.com/anotherhadi/jwt-tui/internal/keys"
	"github.com/anotherhadi/jwt-tui/internal/style"
)

//go:embed docs.md
var jwtDocsMD string

// Panel indices in clockwise order starting top-left:
//
//	top-left (0=JWT) → top-right (1=Header)
//	                         ↓
//	bot-left (3=Secret) ← bot-right (2=Payload)
const (
	panelJWT     = 0
	panelHeader  = 1
	panelPayload = 2
	panelSecret  = 3
)

const exampleJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

var panelPlaceholders = [4]string{
	exampleJWT,
	"{\n  \"alg\": \"HS256\",\n  \"typ\": \"JWT\"\n}",
	"{\n  \"sub\": \"1234567890\",\n  \"name\": \"John Doe\",\n  \"iat\": 1516239022\n}",
	"your-256-bit-secret",
}

var panelTAPlaceholders = [4]string{
	exampleJWT,
	"{\n  \"alg\": \"HS256\",\n  \"typ\": \"JWT\"\n}",
	"{\n  \"sub\": \"1234567890\",\n  \"name\": \"John Doe\",\n  \"iat\": 1516239022\n}",
	"your-256-bit-secret",
}

type panelState struct {
	vp      viewport.Model
	ta      textarea.Model
	editing bool
}

type keyMap struct {
	CycleFocus   key.Binding
	Edit         key.Binding
	EditExternal key.Binding
	Clear        key.Binding
	Reset        key.Binding
	Copy         key.Binding
	Paste        key.Binding
	Docs         key.Binding
	HelpToggle   key.Binding
	Quit         key.Binding
	width        int
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CycleFocus, k.Edit, k.EditExternal, k.HelpToggle, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	all := []key.Binding{k.CycleFocus, k.Edit, k.EditExternal, k.Copy, k.Paste, k.Clear, k.Reset, k.Docs, k.Quit}
	return keys.ChunkByWidth(all, k.width)
}

type docsKeyMap struct {
	Close key.Binding
}

func (k docsKeyMap) ShortHelp() []key.Binding  { return []key.Binding{k.Close} }
func (k docsKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{{k.Close}} }

type Model struct {
	panels  [4]panelState
	initial [4]string // per-panel initial values (for reset)
	focus   int

	showDocs bool
	docsVP   viewport.Model

	pendingEditorPanel int
	pendingPastePanel  int

	sigValid  *bool
	sigStatus string
	errMsg    string

	help     help.Model
	keymap   keyMap
	docsKeys docsKeyMap

	width, height int
}

func New(initialToken, initialSecret string) Model {
	token := strings.TrimSpace(initialToken)
	secret := strings.TrimSpace(initialSecret)

	var initVals [4]string
	initVals[panelJWT] = token
	initVals[panelSecret] = secret
	if token != "" {
		header, payload, _ := jwt.Decode(token)
		initVals[panelHeader] = header
		initVals[panelPayload] = payload
	}

	m := Model{
		initial: initVals,
		focus:   panelJWT,
		help:    ilovetui.NewHelp(),
		keymap: keyMap{
			CycleFocus:   keys.Keys.CycleFocus,
			Edit:         keys.Keys.Edit,
			EditExternal: keys.Keys.EditExternal,
			Clear:        keys.Keys.Clear,
			Reset:        keys.Keys.Reset,
			Copy:         keys.Keys.Copy,
			Paste:        keys.Keys.Paste,
			Docs:         keys.Keys.Docs,
			HelpToggle:   keys.Keys.HelpToggle,
			Quit:         keys.Keys.Quit,
		},
		docsKeys: docsKeyMap{
			Close: keys.Keys.Docs,
		},
	}

	for i := range m.panels {
		ta := ilovetui.NewTextarea(false)
		ta.Placeholder = panelTAPlaceholders[i]
		vp := ilovetui.NewViewport()
		vp.SoftWrap = true
		m.panels[i].ta = ta
		m.panels[i].vp = vp
	}

	m.docsVP = ilovetui.NewViewport()
	m.docsVP.SoftWrap = true

	for i, val := range initVals {
		m.panels[i].ta.SetValue(val)
		if val != "" {
			m.setViewportContent(i, val)
		}
	}

	if token != "" {
		m.revalidate()
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) helpHeight() int {
	if !m.help.ShowAll {
		return 1
	}
	max := 0
	for _, col := range m.keymap.FullHelp() {
		if len(col) > max {
			max = len(col)
		}
	}
	return max
}

func (m *Model) setViewportContent(panel int, raw string) {
	var content string
	switch panel {
	case panelHeader, panelPayload:
		content = highlight.JSON(raw)
	case panelJWT:
		content = highlight.JWT(raw)
	default:
		content = lipgloss.NewStyle().Foreground(ilovetui.S.Text).Render(raw)
	}
	m.panels[panel].vp.SetContent(content)
}

func (m *Model) recalcSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	leftW := m.width / 2
	rightW := m.width - leftW
	helpH := m.helpHeight()
	availH := m.height - helpH - 1 // -1 for the error line
	topH := availH / 2
	bottomH := availH - topH

	setPanel := func(idx, w, h int) {
		cw := max(1, w-2)
		ch := max(1, h-2)
		m.panels[idx].vp.SetWidth(cw)
		m.panels[idx].vp.SetHeight(ch)
		m.panels[idx].ta.SetWidth(cw)
		m.panels[idx].ta.SetHeight(ch)
	}

	setPanel(panelJWT, leftW, topH)
	setPanel(panelHeader, rightW, topH)
	setPanel(panelPayload, rightW, bottomH)
	setPanel(panelSecret, leftW, bottomH)

	m.help.SetWidth(m.width)

	docsAvailH := m.height - 1
	m.docsVP.SetHeight(max(1, docsAvailH-2))
	m.docsVP.SetWidth(max(1, m.width-4))
}

func (m *Model) revalidate() {
	jwtVal := m.panels[panelJWT].ta.Value()
	secVal := m.panels[panelSecret].ta.Value()
	if jwtVal == "" {
		m.sigValid = nil
		m.sigStatus = ""
		return
	}
	valid, err := jwt.Verify(jwtVal, secVal)
	if err != nil {
		m.sigValid = nil
		m.sigStatus = ""
		m.errMsg = err.Error()
		return
	}
	m.errMsg = ""
	m.sigValid = &valid
	if valid {
		m.sigStatus = "✓ Signature Verified"
	} else {
		m.sigStatus = "✗ Invalid Signature"
	}
}

func (m *Model) decodeJWT() {
	token := m.panels[panelJWT].ta.Value()
	if token == "" {
		m.sigValid = nil
		m.sigStatus = ""
		m.errMsg = ""
		m.panels[panelHeader].ta.SetValue("")
		m.panels[panelHeader].vp.SetContent("")
		m.panels[panelPayload].ta.SetValue("")
		m.panels[panelPayload].vp.SetContent("")
		return
	}
	header, payload, err := jwt.Decode(token)
	if err != nil {
		m.sigValid = nil
		m.sigStatus = ""
		m.errMsg = err.Error()
		m.panels[panelHeader].ta.SetValue("")
		m.panels[panelHeader].vp.SetContent("")
		m.panels[panelPayload].ta.SetValue("")
		m.panels[panelPayload].vp.SetContent("")
		return
	}
	m.errMsg = ""
	m.panels[panelHeader].ta.SetValue(header)
	m.panels[panelPayload].ta.SetValue(payload)
	m.setViewportContent(panelHeader, header)
	m.setViewportContent(panelPayload, payload)
	m.revalidate()
}

func (m *Model) rebuildJWT() {
	header := m.panels[panelHeader].ta.Value()
	payload := m.panels[panelPayload].ta.Value()
	if header == "" && payload == "" {
		m.panels[panelJWT].ta.SetValue("")
		m.panels[panelJWT].vp.SetContent("")
		m.sigValid = nil
		m.sigStatus = ""
		m.errMsg = ""
		return
	}
	token, err := jwt.Encode(header, payload, m.panels[panelSecret].ta.Value())
	if err != nil {
		m.sigValid = nil
		m.sigStatus = ""
		m.errMsg = err.Error()
		return
	}
	m.errMsg = ""
	m.panels[panelJWT].ta.SetValue(token)
	m.setViewportContent(panelJWT, token)
	m.revalidate()
}

func (m *Model) exitEditMode() {
	p := &m.panels[m.focus]
	if !p.editing {
		return
	}
	p.editing = false
	p.ta.Placeholder = panelTAPlaceholders[m.focus]
	p.ta.Blur()
	raw := p.ta.Value()
	if raw != "" {
		m.setViewportContent(m.focus, raw)
	}
	switch m.focus {
	case panelHeader, panelPayload:
		m.rebuildJWT()
	case panelJWT:
		m.decodeJWT()
	case panelSecret:
		m.revalidate()
	}
}

func (m *Model) enterEditMode() tea.Cmd {
	p := &m.panels[m.focus]
	p.editing = true
	p.ta.Placeholder = ""
	return p.ta.Focus()
}

// clearPanel empties the focused panel and enters edit mode.
func (m *Model) clearPanel() tea.Cmd {
	m.exitEditMode()
	m.panels[m.focus].ta.SetValue("")
	m.panels[m.focus].vp.SetContent("")
	// Clearing JWT also wipes derived header/payload
	if m.focus == panelJWT {
		m.panels[panelHeader].ta.SetValue("")
		m.panels[panelHeader].vp.SetContent("")
		m.panels[panelPayload].ta.SetValue("")
		m.panels[panelPayload].vp.SetContent("")
		m.sigValid = nil
		m.sigStatus = ""
	} else {
		switch m.focus {
		case panelHeader, panelPayload:
			m.rebuildJWT()
		case panelSecret:
			m.revalidate()
		}
	}
	return m.enterEditMode()
}

// resetPanel restores the focused panel to its initial value.
func (m *Model) resetPanel() {
	m.exitEditMode()
	val := m.initial[m.focus]
	m.panels[m.focus].ta.SetValue(val)
	if val != "" {
		m.setViewportContent(m.focus, val)
	} else {
		m.panels[m.focus].vp.SetContent("")
	}
	// Resetting JWT also restores derived header/payload from initial
	if m.focus == panelJWT {
		m.panels[panelHeader].ta.SetValue(m.initial[panelHeader])
		if m.initial[panelHeader] != "" {
			m.setViewportContent(panelHeader, m.initial[panelHeader])
		} else {
			m.panels[panelHeader].vp.SetContent("")
		}
		m.panels[panelPayload].ta.SetValue(m.initial[panelPayload])
		if m.initial[panelPayload] != "" {
			m.setViewportContent(panelPayload, m.initial[panelPayload])
		} else {
			m.panels[panelPayload].vp.SetContent("")
		}
		m.revalidate()
	} else {
		switch m.focus {
		case panelHeader, panelPayload:
			m.rebuildJWT()
		case panelSecret:
			m.revalidate()
		}
	}
}

func (m *Model) renderDocs() {
	width := max(40, m.docsVP.Width())
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(ilovetui.GlamourStyleConfig()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		m.docsVP.SetContent(jwtDocsMD)
		return
	}
	rendered, err := renderer.Render(jwtDocsMD)
	if err != nil {
		m.docsVP.SetContent(jwtDocsMD)
		return
	}
	m.docsVP.SetContent(rendered)
	m.docsVP.SetYOffset(0)
}

func (m Model) borderFor(panel int) lipgloss.Style {
	if panel == m.focus && !m.showDocs {
		return ilovetui.S.PanelFocused
	}
	return ilovetui.S.Panel
}

func (m Model) panelTitle(panel int, name string) string {
	bc := lipgloss.NewStyle().Foreground(m.borderFor(panel).GetBorderTopForeground())
	title := bc.Render(name)
	if panel == m.focus && m.panels[panel].editing {
		title += ilovetui.S.Faint.Render(" [edit]")
	}
	return title
}

func (m Model) secretTitle() string {
	name := m.panelTitle(panelSecret, "Secret")

	var sigStr string
	if m.sigValid == nil {
		sigStr = ilovetui.S.Faint.Render("·")
	} else if *m.sigValid {
		sigStr = lipgloss.NewStyle().Foreground(ilovetui.S.Success).Render("✓")
	} else {
		sigStr = lipgloss.NewStyle().Foreground(ilovetui.S.Error).Render("✗")
	}

	return name + ilovetui.S.Faint.Render(" · ") + sigStr
}

func (m *Model) renderPanelContent(panel int) string {
	p := &m.panels[panel]
	if p.editing {
		return p.ta.View()
	}
	if p.ta.Value() == "" {
		return ilovetui.S.Faint.Render(panelPlaceholders[panel])
	}
	return ilovetui.ViewportView(&p.vp)
}

func (m *Model) renderPanel(panel int, title string, w, h int) string {
	return style.RenderWithTitle(m.borderFor(panel), title, m.renderPanelContent(panel), w, h)
}
