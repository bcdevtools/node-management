package node

import (
	"encoding/json"
	"fmt"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func GetPruneAddrBookCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "prune-addrbook [addrbook.json]",
		Short: "Prune " + addrBookFileName + " and keep only live peers",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			inputFilePath := strings.TrimSpace(args[0])

			if inputFilePath == "" {
				utils.ExitWithErrorMsg("ERR: required input file")
				return
			}

			addrBook, err := readAddrBook(inputFilePath)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to read address book:", err)
				return
			}

			livePeers := addrBook.GetLivePeers(48 * time.Hour)

			newAddrBook := types.AddrBook{
				Key:   addrBook.Key,
				Addrs: livePeers,
			}

			bz, err := json.MarshalIndent(newAddrBook, "", "  ")
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to marshal JSON new address book:", err)
				return
			}

			backupFile := inputFilePath + ".bak"
			err = os.WriteFile(backupFile, bz, 0644)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to write backup file:", err)
				return
			}
			fmt.Println("Backup file:", backupFile)

			err = os.WriteFile(inputFilePath, bz, 0644)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to re-write file:", err)
				return
			}

			if len(livePeers) > 0 {
				fmt.Println("Pruned successfully, keep", len(livePeers), "live-peers.")
			} else {
				fmt.Println("Pruned successfully but no peer left.")
			}
		},
	}

	return cmd
}
