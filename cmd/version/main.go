package version

import (
    "fmt"
    "github.com/spf13/cobra"
)

// buildTime will be set via ldflags at build time. Default to "unknown".
var buildTime = "unknown"

func NewCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print the build version/timestamp",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("gbr version built at %s\n", buildTime)
        },
    }
}
