package cmd

import (
	"github.com/bcdevtools/node-management/cmd/keys"
	"github.com/bcdevtools/node-management/cmd/node"
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   constants.BINARY_NAME,
	Short: constants.APP_DESC,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		type report struct {
			message    string
			suggestion string
		}
		var reports []report
		if !utils.HasBinaryName("aria2c") {
			reports = append(reports, report{
				message:    "Require `aria2c` installed!",
				suggestion: "sudo apt install -y aria2",
			})
		}
		if !utils.HasBinaryName("lz4") {
			reports = append(reports, report{
				message:    "Require `lz4` installed!",
				suggestion: "sudo apt install snapd -y && sudo snap install lz4",
			})
		}
		if !utils.HasBinaryName("jq") {
			reports = append(reports, report{
				message:    "Require `jq` installed!",
				suggestion: "sudo apt install -y jq",
			})
		}
		if !utils.HasBinaryName("ssh-keygen") {
			reports = append(reports, report{
				message:    "Require `ssh-keygen` installed!",
				suggestion: "sudo apt install -y openssh-client",
			})
		}
		if !utils.HasBinaryName("rsync") {
			reports = append(reports, report{
				message:    "Require `rsync` installed!",
				suggestion: "sudo apt install -y rsync",
			})
		}

		if len(reports) < 1 {
			return
		}

		reports = append(reports, report{
			message: "Please install the required tools before running the command",
		})
		for i, r := range reports {
			if i > 0 {
				utils.PrintfStdErr("\n\n")
			}
			utils.PrintlnStdErr(r.message)
			if len(r.suggestion) > 0 {
				utils.PrintlnStdErr(" " + r.suggestion)
			}
		}
		os.Exit(1)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true    // hide the 'completion' subcommand
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true}) // hide the 'help' subcommand

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(node.GetNodeCommands())
	rootCmd.AddCommand(keys.GetKeysCommands())
}
