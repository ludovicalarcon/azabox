package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/installer"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
	"gitlab.com/ludovic-alarcon/azabox/internal/state"
)

const (
	InstallUseMessage   = "install"
	InstallShortMessage = "install binary for current user"

	DefaultBinaryVersion = "latest"

	ArgsCountErrorMessage = "install need at least one argument, see above usage"
)

type InstallConfig struct {
	azaInstaller installer.Installer
	azaState     state.State
}

func createBinaryInfo(binaryName, version string) dto.BinaryInfo {
	binaryInfo := dto.BinaryInfo{
		FullName: binaryName,
		Version:  version,
	}

	info := strings.Split(binaryInfo.FullName, "/")
	if len(info) == 1 {
		binaryInfo.Name = binaryInfo.FullName
		binaryInfo.Owner = binaryInfo.FullName
	} else {
		binaryInfo.Owner = info[0]
		binaryInfo.Name = info[1]
	}

	binaryInfo.FullName = fmt.Sprintf("%s/%s", binaryInfo.Owner, binaryInfo.Name)

	return binaryInfo
}

func binariesInfoFromArgs(installArgs []string, version string) []dto.BinaryInfo {
	binaryInfosSlice := make([]dto.BinaryInfo, 0, len(installArgs))

	for _, binaryName := range installArgs {
		binaryInfo := createBinaryInfo(binaryName, version)
		binaryInfosSlice = append(binaryInfosSlice, binaryInfo)
	}

	return binaryInfosSlice
}

func installBinary(binaryInfo *dto.BinaryInfo, cfg InstallConfig) error {
	logging.Logger.Debugw("Installing binary", "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.Version)
	fmt.Printf("Installing binary \"%s\" with version \"%s\"\n", binaryInfo.FullName, binaryInfo.Version)

	err := cfg.azaState.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if cfg.azaState.Has(binaryInfo.FullName) {
		return errors.New("binary already installed, use update command to download newer version")
	}

	resolvers := resolver.GetRegistryResolver().GetResolvers()
	resolvedUrl := ""
	for resolver := range resolvers {
		url, err := resolver.Resolve(binaryInfo)
		if err == nil && url != "" {
			resolvedUrl = url
			logging.Logger.Debugw("Matched resolver", "type",
				fmt.Sprintf("%T", resolver), "url", url)
			break
		}
	}

	if resolvedUrl == "" {
		logging.Logger.Debugw("Binary not found", "binary", binaryInfo.Name, "owner",
			binaryInfo.Owner, "version", binaryInfo.Version)
		fmt.Printf("Binary \"%s\" with version \"%s\" not found\n", binaryInfo.FullName, binaryInfo.Version)
		return cfg.azaState.Save()
	}

	if err = cfg.azaInstaller.Install(binaryInfo, resolvedUrl); err != nil {
		return err
	}

	cfg.azaState.UpdateEntrie(*binaryInfo)
	return cfg.azaState.Save()
}

func newInstallCommand(localInstaller installer.Installer, localState state.State) *cobra.Command {
	var version string
	cfg := InstallConfig{
		azaInstaller: localInstaller,
		azaState:     localState,
	}

	cmd := &cobra.Command{
		Use:   InstallUseMessage,
		Short: InstallShortMessage,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return errors.New(ArgsCountErrorMessage)
			}

			binaryInfoSlice := binariesInfoFromArgs(args, version)
			for _, binaryInfo := range binaryInfoSlice {
				err := installBinary(&binaryInfo, cfg)
				if err != nil {
					return err
				}
			}

			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringVarP(&version, "version", "v",
		DefaultBinaryVersion, "desired version of the binary")

	return cmd
}
