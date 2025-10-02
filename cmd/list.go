package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/ludovic-alarcon/azabox/internal/state"
)

const (
	ListUseMessage   = "list"
	ListShortMessage = "list installed binaries for current user"
)

func newListCommand(azaState state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   ListUseMessage,
		Short: ListShortMessage,
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := executeListCommand(azaState)
			fmt.Println(list)
			return err
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func executeListCommand(azaState state.State) (string, error) {
	err := azaState.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	var sb strings.Builder
	if len(azaState.Entries()) == 0 {
		sb.WriteString("No binary installed\n")
	} else {
		sb.WriteString("Binaries installed:\n")
		for _, binary := range azaState.Entries() {
			sb.WriteString(fmt.Sprintf("- %s\n", binary.String()))
		}
	}
	return sb.String(), nil
}
