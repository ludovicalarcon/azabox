package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
)

const (
	InstallUseMessage   = "install"
	InstallShortMessage = "Install binary for current user"

	DefaultBinaryVersion = "latest"
)

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

func installBinary(binaryInfo dto.BinaryInfo) error {
	logging.Logger.Debugw("Installing binary", "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.Version)
	fmt.Printf("Installing binary \"%s\" with version \"%s\"\n", binaryInfo.FullName, binaryInfo.Version)

	resolvers := resolver.GetRegistryResolver().GetResolvers()
	for _, resolver := range resolvers {
		_, err := resolver.Resolve(binaryInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func newInstallCommand() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   InstallUseMessage,
		Short: InstallShortMessage,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("install need at least one argument, see above usage")
			}

			binaryInfoSlice := binariesInfoFromArgs(args, version)
			for _, binaryInfo := range binaryInfoSlice {
				err := installBinary(binaryInfo)
				if err != nil {
					return err
				}
			}

			return nil
		},
		SilenceErrors: true,
	}

	cmd.Flags().StringVarP(&version, "version", "v",
		DefaultBinaryVersion, "desired version of the binary")

	return cmd
}
