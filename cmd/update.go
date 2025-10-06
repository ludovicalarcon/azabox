package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/installer"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
	"gitlab.com/ludovic-alarcon/azabox/internal/state"
)

const (
	UpdateUseMessage   = "update"
	UpdateShortMessage = "update installed binaries for current user"
)

type UpdateCommandConfig struct {
	azaInstaller installer.Installer
	azaState     state.State
}

func newUpdateCommand(azaInstaller installer.Installer, azaState state.State) *cobra.Command {
	cfg := UpdateCommandConfig{
		azaInstaller: azaInstaller,
		azaState:     azaState,
	}

	cmd := &cobra.Command{
		Use:   UpdateUseMessage,
		Short: UpdateShortMessage,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeUpdateCommand(cfg, args...)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func executeUpdateCommand(cfg UpdateCommandConfig, args ...string) error {
	err := cfg.azaState.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if len(args) > 0 {
		for _, binaryName := range args {
			binaryInfo, ok := cfg.azaState.Entry(dto.NormalizeName(binaryName))
			if !ok {
				return fmt.Errorf("binary %s is not installed (or not managed by azabox)", binaryName)
			}
			if err := checkUpdate(binaryInfo, cfg); err != nil {
				return err
			}
		}
	} else {
		for _, binaryInfo := range cfg.azaState.Entries() {
			if err := checkUpdate(binaryInfo, cfg); err != nil {
				return err
			}
		}
	}

	return cfg.azaState.Save()
}

func checkUpdate(binaryInfo dto.BinaryInfo, cfg UpdateCommandConfig) error {
	version, lresolver, err := resolveLatestVersion(binaryInfo)
	if err != nil {
		return err
	}

	logging.Logger.Debugw("update command", "resolvedVersion", version,
		"name", binaryInfo.DisplayName(), "currentVersion", binaryInfo.InstalledVersion)
	if version != binaryInfo.InstalledVersion {
		logging.Logger.Debugw("update binary", "name", binaryInfo.DisplayName(),
			"currentVersion", binaryInfo.InstalledVersion, "newVersion", version)
		fmt.Printf("Updating %s from %s to %s\n", binaryInfo.DisplayName(),
			binaryInfo.InstalledVersion, version)

		return update(lresolver, &binaryInfo, cfg)
	} else {
		fmt.Printf("Binary %s is up to date\n", binaryInfo.DisplayName())
	}
	return nil
}

func resolveLatestVersion(binaryInfo dto.BinaryInfo) (string, resolver.Resolver, error) {
	resolvers := resolver.GetRegistryResolver().GetResolvers()
	for lresolver := range resolvers {
		if lresolver.Name() == binaryInfo.Resolver {
			version, err := lresolver.ResolveLatestVersion(binaryInfo)
			return version, lresolver, err
		}
	}
	return "", nil, fmt.Errorf("unknown resolver %s", binaryInfo.Resolver)
}

func update(lresolver resolver.Resolver, binaryInfo *dto.BinaryInfo, cfg UpdateCommandConfig) error {
	binaryInfo.Version = resolver.LatestVersion
	resolvedUrl, err := lresolver.Resolve(binaryInfo)
	if err != nil {
		return err
	}

	if err = cfg.azaInstaller.Install(binaryInfo, resolvedUrl); err != nil {
		return err
	}

	cfg.azaState.UpdateEntrie(*binaryInfo)
	return nil
}
