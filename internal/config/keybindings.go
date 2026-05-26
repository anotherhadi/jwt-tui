package config

type Keybindings struct {
	Quit         string `mapstructure:"quit"`
	CycleFocus   string `mapstructure:"cycle_focus"`
	Edit         string `mapstructure:"edit"`
	EditExternal string `mapstructure:"edit_external"`
	Docs         string `mapstructure:"docs"`
	HelpToggle   string `mapstructure:"help_toggle"`
	Clear        string `mapstructure:"clear"`
	Reset        string `mapstructure:"reset"`
	Copy         string `mapstructure:"copy"`
	Paste        string `mapstructure:"paste"`
}
