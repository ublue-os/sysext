package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/cmd/layer"
	"github.com/ublue-os/sysext/cmd/mount"
	"github.com/ublue-os/sysext/internal"
	appLogging "github.com/ublue-os/sysext/pkg/logging"
)

var RootCmd = &cobra.Command{
	Use:          "bext",
	Short:        "Manager for Systemd system extensions",
	Long:         `Manage your systemd system extensions from your CLI, managing their cache, multiple versions, and building.`,
	SilenceUsage: true,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	fLogFile   *string
	fLogLevel  *string
	fNoLogging *bool
)

func init() {
	fLogFile = RootCmd.PersistentFlags().String("log-file", "", "File where user facing logs will be written to")
	fLogLevel = RootCmd.PersistentFlags().String("log-level", "info", "File where user facing logs will be written to")
	fNoLogging = RootCmd.PersistentFlags().Bool("quiet", false, "Do not log anything to anywhere")
	internal.Config.NoProgress = RootCmd.PersistentFlags().Bool("no-progress", false, "Do not use progress bars whenever they would be")

	RootCmd.AddCommand(layer.LayerCmd)
	RootCmd.AddCommand(mount.MountCmd)
	RootCmd.AddCommand(AddToPathCmd)

	if *fNoLogging {
		slog.SetDefault(appLogging.NewMuteLogger())
		return
	}
	var logWriter *os.File = os.Stdout
	if *fLogFile != "" {
		abs, err := filepath.Abs(path.Clean(*fLogFile))
		if err != nil {
			os.Exit(1)
		}
		logWriter, err = os.Create(abs)
		if err != nil {
			fmt.Println("Could not open log file")
			os.Exit(1)
		}
		defer logWriter.Close()
	}

	logLevel, err := appLogging.StrToLogLevel(*fLogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	main_app_logger := slog.New(appLogging.SetupAppLogger(logWriter, logLevel, *fLogFile != ""))

	slog.SetDefault(main_app_logger)

}
