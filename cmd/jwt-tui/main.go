package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/jwt-tui/internal/config"
	"github.com/anotherhadi/jwt-tui/internal/keys"
	"github.com/anotherhadi/jwt-tui/internal/ui"
	"github.com/spf13/pflag"
)

// Version is overwritten at build time by goreleaser/ldflag with the current version tag, or "dev" if not set.
var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

func main() {
	var (
		flagConfig           = pflag.StringP("config", "c", "", "path to config file")
		flagAddDefaultConfig = pflag.Bool("add-default-config", false, "copy the default config file to the config path and exit")
		flagToken            = pflag.StringP("token", "t", "", "pre-fill the encoded JWT token")
		flagSecret           = pflag.StringP("secret", "s", "", "pre-fill the secret key")
		flagVersion          = pflag.BoolP("version", "v", false, "print version")
	)
	pflag.CommandLine.SetOutput(os.Stdout)
	pflag.Usage = func() {
		fmt.Println("Usage: jwt-tui [flags]")
		fmt.Println("")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// Accept a bare positional argument as the token
	if *flagToken == "" && pflag.NArg() > 0 {
		t := pflag.Arg(0)
		flagToken = &t
	}

	home, _ := os.UserHomeDir()
	cfgPath := filepath.Join(home, ".config", "jwt-tui", "config.yaml")
	if *flagConfig != "" {
		cfgPath = *flagConfig
	}

	if *flagAddDefaultConfig {
		if err := config.WriteDefaultConfig(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "write-config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("default config written to %s\n", cfgPath)
		return
	}

	if err := config.Load(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	keys.Init(config.Global)

	m := ui.New(*flagToken, *flagSecret)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
