package nodedaemon

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node-daemon",
		Short: "Run the node daemon (inside VMs)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("node-daemon not yet implemented")
		},
	}

	return cmd
}