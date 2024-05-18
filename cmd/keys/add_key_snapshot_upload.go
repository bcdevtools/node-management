package keys

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

func GetAddSnapshotUploadKeyCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "add-snapshot-upload-ssh-key",
		Aliases: []string{"ss"},
		Short:   "Add snapshot upload SSH key",
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			fmt.Println("Chain name:")
			chainName := utils.ReadText(false)

			chainNameNormalized := strings.ToLower(chainName)
			chainNameNormalized = strings.ReplaceAll(chainNameNormalized, " ", "_")
			chainNameNormalized = strings.ReplaceAll(chainNameNormalized, "-", "_")

			fmt.Println("Network type:")
			fmt.Println("1. Mainnet")
			fmt.Println("2. Testnet")
			networkTypeOption := utils.ReadNumber(1, 2)

			var sb strings.Builder
			sb.WriteString("id_snapshot_upload_")
			sb.WriteString(chainNameNormalized)
			if networkTypeOption == 1 {
				sb.WriteString("_mainnet")
			} else {
				sb.WriteString("_testnet")
			}
			sb.WriteString("_")
			sb.WriteString(strings.ReplaceAll(time.Now().UTC().Format(time.DateOnly), "-", "_"))

			id := sb.String()

			addSshKey(id)
		},
	}

	return cmd
}
