package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/state"
)

const (
	ListUseMessage   = "list"
	ListShortMessage = "list installed binaries for current user"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ListUseMessage,
		Short: ListShortMessage,
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := executeListCommand()
			fmt.Println(list)
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func executeListCommand() (string, error) {
	azaState := state.NewState(filepath.Clean(
		filepath.Join(state.StateDirectory(), state.StateFileName)))
	err := azaState.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	var sb strings.Builder
	if len(azaState.Binaries) == 0 {
		sb.WriteString("No binary installed\n")
	} else {
		sb.WriteString("Binaries installed:\n")
		for _, binary := range azaState.Binaries {
			sb.WriteString(fmt.Sprintf("- %s\n", binary.String()))
		}
	}
	return sb.String(), nil
}
