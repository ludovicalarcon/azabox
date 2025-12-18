package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/installer"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
	"gitlab.com/ludovic-alarcon/azabox/internal/state"
)

const (
	RootUseMessage   = "azabox"
	RootShortMessage = "azabox - A per-user binary manager"
	RootLongMessage  = `azabox is a CLI tool to install, manage and switch between different versions
of command-line binaries from GitHub, GitLab, or custom URLs.`
)

var rootCmd = &cobra.Command{
	Use:   RootUseMessage,
	Short: RootShortMessage,
	Long:  RootLongMessage,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logging.InitLogger()
	},
}

func setupCommands() error {
	azaInstaller, err := installer.New()
	if err != nil {
		return err
	}
	azaState := state.NewState(filepath.Clean(
		filepath.Join(state.StateDirectory(), state.StateFileName)))

	rootCmd.AddCommand(newInstallCommand(azaInstaller, azaState))
	rootCmd.AddCommand(newListCommand(azaState))
	rootCmd.AddCommand(newUpdateCommand(azaInstaller, azaState))

	return nil
}

func Execute() error {
	defer func() {
		if logging.Logger() != nil {
			logging.Logger().Sync()
		}
	}()

	if err := setupCommands(); err != nil {
		return err
	}

	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", "",
		"Set the logging level (debug, info, warn, error)")

	// initialize the registry with default resolvers
	_ = resolver.GetRegistryResolver().WithDefaultResolvers()
}
