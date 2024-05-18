package node

//goland:noinspection GoSnakeCaseUsage
import (
	setup_check "github.com/bcdevtools/node-management/cmd/node/setup-check"
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

	return cmd

}
