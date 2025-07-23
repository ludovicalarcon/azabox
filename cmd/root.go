package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
)

var rootCmd = &cobra.Command{
	Use:   "azabox",
	Short: "azabox - A per-user binary manager",
	Long: `azabox is a CLI tool to install, manage and switch between different versions
of command-line binaries from GitHub, GitLab, or custom URLs.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		logging.Logger.Debugln("run root command")
		cmd.Help()
		return nil
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.InitLogger(logging.Config{Encoding: logging.Console})
	},
}

func Execute() error {
	defer func() {
		if logging.Logger != nil {
			logging.Logger.Sync()
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		logging.Logger.Errorw("command exited with error", err)
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", "",
		"Set the logging level (debug, info, warn, error)")
}
