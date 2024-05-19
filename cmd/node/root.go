package node

//goland:noinspection GoSnakeCaseUsage
import (
	setup_check "github.com/bcdevtools/node-management/cmd/node/setup-check"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
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
	err := validation.PossibleNodeHome(nodeHomeDirectory)
	if err != nil {
		utils.ExitWithErrorMsg("ERR: invalid node home directory:", err)
		return
	}
}
