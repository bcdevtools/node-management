package keys

//goland:noinspection GoSnakeCaseUsage
import (
	"github.com/spf13/cobra"
)

func GetKeysCommands() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "keys",
		Short: "Manage keys",
	}

	cmd.AddCommand(GetAddSnapshotUploadKeyCmd())

	return cmd
}
