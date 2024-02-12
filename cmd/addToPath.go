package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
)

var AddToPathCmd = &cobra.Command{
	Use:   "add-to-path",
	Short: "Add the mounted layer binaries to your path",
	Long:  `Write a snippet for your shell of the mounted path for the activated sysext layers`,
	RunE:  addToPathCmd,
}

type ShellDefinition struct {
	Snippet string
	RcPath  string
}

var (
	fPathPath *string
	fRCPath   *string
)

func init() {
	fPathPath = AddToPathCmd.Flags().StringP("path", "p", "/tmp/extensions.d/bin", "Path where all shared binaries are being mounted to")
	fRCPath = AddToPathCmd.Flags().StringP("rc-path", "r", "", "RC path for your chosen shell instead of the default")
}

func addToPathCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return internal.NewPositionalError("SHELL")
	}

	user_home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	user_config, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	var defaultValues = map[string]ShellDefinition{
		"bash": {
			RcPath:  fmt.Sprintf("%s/.bashrc", user_home),
			Snippet: fmt.Sprintf("[ -e %s ] && PATH=\"$PATH:%s\" \n", *fPathPath, *fPathPath),
		},
		"zsh": {
			RcPath:  fmt.Sprintf("%s/.zshrc", user_home),
			Snippet: fmt.Sprintf("[ -e %s ] && PATH=\"$PATH:%s\" \n", *fPathPath, *fPathPath),
		},
		"nu": {
			RcPath:  fmt.Sprintf("%s/config.nu", user_config),
			Snippet: fmt.Sprintf("$env.PATH = ($env.PATH | split row (char esep) | append %s)\n", *fPathPath),
		},
	}

	var valid_stuff []string
	for key := range defaultValues {
		valid_stuff = append(valid_stuff, key)
	}

	if !slices.Contains(valid_stuff, args[0]) {
		slog.Warn(fmt.Sprintf("Could not find shell %s, valid shells are: %s", args[0], strings.Join(valid_stuff, ", ")))
		os.Exit(1)
	}

	var rcPath string
	if *fRCPath != "" {
		rcPath = path.Clean(*fRCPath)
	} else {
		rcPath = defaultValues[args[0]].RcPath
	}

	if _, err := fileio.FileAppendS(rcPath, defaultValues[args[0]].Snippet); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Successfully written snippet to %s\n", rcPath))

	return nil
}
