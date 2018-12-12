package commands

import (
	"os"

	"github.com/spf13/cobra"

	"cloud-mta-build-tool/internal/tpl"
)

var initModeFlag string
var descriptorInitFlag string
var sourceInitFlag string
var targetInitFlag string

func init() {
	initProcessCmd.Flags().StringVarP(&initModeFlag, "mode", "m", "", "Mode of Makefile generation - default/verbose")
	initProcessCmd.Flags().StringVarP(&descriptorInitFlag, "desc", "d", "", "Descriptor MTA - dev/dep")
	initProcessCmd.Flags().StringVarP(&sourceInitFlag, "source", "s", "", "Provide MTA source")
	initProcessCmd.Flags().StringVarP(&targetInitFlag, "target", "t", "", "Provide MTA target")
}

var initProcessCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate Makefile",
	Long:  "Generate Makefile as manifest which describe's the build process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Generate build script
		err := tpl.ExecuteMake(sourceInitFlag, targetInitFlag, descriptorInitFlag, initModeFlag, os.Getwd)
		logError(err)
	},
}