package node

//goland:noinspection GoSnakeCaseUsage
import (
	setup_check "github.com/bcdevtools/node-management/cmd/node/setup-check"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
)

func GetNodeCommands() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "node",
		Short: "Manage nodes",
	}

	cmd.AddCommand(setup_check.GetStepCheckCmd())
	cmd.AddCommand(GetExtractAddrBookCmd())
	cmd.AddCommand(GetPruneAddrBookCmd())
	cmd.AddCommand(GetPruneNodeDataCmd())
	cmd.AddCommand(GetStateSyncCmd())
	cmd.AddCommand(GetZipSnapshotCmd())

	return cmd
}

func validateNodeHomeDirectory(nodeHomeDirectory string) {
	if nodeHomeDirectory == "" {
		utils.ExitWithErrorMsg("ERR: required node home directory")
		return
	}
	_, exists, isDir, err := utils.FileInfo(nodeHomeDirectory)
	if err != nil {
		utils.ExitWithErrorMsg("ERR: failed to check node home directory:", err)
		return
	}
	if !exists {
		utils.ExitWithErrorMsg("ERR: node home directory does not exist:", nodeHomeDirectory)
		return
	}
	if !isDir {
		utils.ExitWithErrorMsg("ERR: specified path is not a directory:", nodeHomeDirectory)
		return
	}
}
