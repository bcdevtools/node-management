package node

import (
	"encoding/json"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"time"
)

const addrBookFileName = "addrbook.json"

func GetExtractAddrBookCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "extract-addrbook [input-file] [output-file]",
		Short: "Extract live-peers from " + addrBookFileName,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			inputFilePath := strings.TrimSpace(args[0])
			outputFilePath := strings.TrimSpace(args[1])

			if inputFilePath == "" || outputFilePath == "" {
				utils.ExitWithErrorMsg("ERR: input and output file names are required")
				return
			}

			dirInput, fileName := path.Split(inputFilePath)
			if fileName != addrBookFileName {
				utils.ExitWithErrorMsg("ERR: input file name must be " + addrBookFileName)
				return
			}
			dirOutput, _ := path.Split(outputFilePath)
			if dirInput == dirOutput {
				utils.ExitWithErrorMsg("ERR: input and output files must be different")
				return
			}
			if !strings.HasPrefix(dirInput, "/") {
				utils.ExitWithErrorMsg("ERR: input file path must be absolute")
				return
			}
			if !strings.HasPrefix(dirOutput, "/") {
				utils.ExitWithErrorMsg("ERR: output file path must be absolute")
				return
			}
			_, exists, _, err := utils.FileInfo(dirOutput)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check output directory:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: output directory does not exist:", dirOutput)
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

			err = os.WriteFile(outputFilePath, bz, 0644)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to write file:", err)
				return
			}
		},
	}

	return cmd
}
