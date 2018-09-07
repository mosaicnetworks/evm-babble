package commands

import (
	"fmt"

	"github.com/mosaicnetworks/babble/version"
	"github.com/spf13/cobra"
)

// VersionCmd displays the version of the evm-babble program being used
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}
