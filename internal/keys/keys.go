package keys

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/jwt-tui/internal/config"
)

type KeyMap struct {
	Quit         key.Binding
	CycleFocus   key.Binding
	Edit         key.Binding
	EditExternal key.Binding
	Docs         key.Binding
	HelpToggle   key.Binding
	Clear        key.Binding
	Reset        key.Binding
	Copy         key.Binding
	Paste        key.Binding
}

var Keys *KeyMap

func Init(cfg *config.Config) {
	kb := cfg.Keybindings
	Keys = &KeyMap{
		Quit:         binding(kb.Quit, "quit"),
		CycleFocus:   binding(kb.CycleFocus, "cycle focus"),
		Edit:         binding(kb.Edit, "edit"),
		EditExternal: binding(kb.EditExternal, "edit in $EDITOR"),
		Docs:         binding(kb.Docs, "docs"),
		HelpToggle:   binding(kb.HelpToggle, "help"),
		Clear:        binding(kb.Clear, "clear"),
		Reset:        binding(kb.Reset, "reset"),
		Copy:         binding(kb.Copy, "copy"),
		Paste:        binding(kb.Paste, "paste"),
	}
}

func parseKeys(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if k := strings.TrimSpace(p); k != "" {
			out = append(out, k)
		}
	}
	return out
}

func ChunkByWidth(bindings []key.Binding, termWidth int) [][]key.Binding {
	cols := termWidth / 26
	if cols < 2 {
		cols = 2
	} else if cols > 7 {
		cols = 7
	}
	perCol := (len(bindings) + cols - 1) / cols
	var out [][]key.Binding
	for i := 0; i < len(bindings); i += perCol {
		end := i + perCol
		if end > len(bindings) {
			end = len(bindings)
		}
		out = append(out, bindings[i:end])
	}
	return out
}

func binding(s, help string) key.Binding {
	keys := parseKeys(s)
	display := strings.Join(keys, "/")
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(display, help))
}
