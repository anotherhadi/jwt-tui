package util

import (
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

type EditorFinishedMsg struct {
	Content string
	Err     error
}

func OpenExternalEditor(content string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	f, err := os.CreateTemp("", "jwt-tui-*.json")
	if err != nil {
		return func() tea.Msg { return EditorFinishedMsg{Err: err} }
	}
	tmpPath := f.Name()
	if _, werr := f.WriteString(content); werr != nil {
		f.Close()
		os.Remove(tmpPath)
		return func() tea.Msg { return EditorFinishedMsg{Err: werr} }
	}
	f.Close()

	return tea.ExecProcess(exec.Command(editor, tmpPath), func(err error) tea.Msg {
		defer os.Remove(tmpPath)
		if err != nil {
			return EditorFinishedMsg{Err: err}
		}
		data, readErr := os.ReadFile(tmpPath)
		if readErr != nil {
			return EditorFinishedMsg{Err: readErr}
		}
		return EditorFinishedMsg{Content: string(data)}
	})
}
