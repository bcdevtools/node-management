package node

import (
	"fmt"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"time"
)

func GetZipSnapshotCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "zip-snapshot [node_home]",
		Short: "Zip node data for snapshot",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			nodeHomeDirectory := strings.TrimSpace(args[0])
			validateNodeHomeDirectory(nodeHomeDirectory)

			dataDirPath := path.Join(nodeHomeDirectory, "data")
			_, exists, isDir, err := utils.FileInfo(dataDirPath)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check data directory:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: data directory does not exist:", dataDirPath)
				return
			}
			if !isDir {
				utils.ExitWithErrorMsg("ERR: required data dir is not a directory:", dataDirPath)
				return
			}

			_, exists, _, err = utils.FileInfo(path.Join(dataDirPath, "application.db"))
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check application.db dir:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: ", dataDirPath, "does not contains data")
				return
			}

			privValStateJsonFilePath := path.Join(dataDirPath, "priv_validator_state.json")
			_, exists, _, err = utils.FileInfo(privValStateJsonFilePath)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check priv_validator_state.json file:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: priv_validator_state.json file does not exist:", privValStateJsonFilePath)
				return
			}
			pvs := &types.PrivateValidatorState{}
			err = pvs.LoadFromJSONFile(privValStateJsonFilePath)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to load priv_validator_state.json file:", err)
				return
			}
			if !pvs.IsEmpty() {
				utils.ExitWithErrorMsg("ERR: priv_validator_state.json file is not empty")
				return
			}

			_ = os.RemoveAll(path.Join(dataDirPath, "snapshots"))
			_ = os.RemoveAll(path.Join(dataDirPath, "tx_index.db"))

			// zip data dir
			zipFileNameWithoutExt := fmt.Sprintf("snapshot_%s.tar.lz4", utils.GetDateTimeStringCompatibleWithFileName(time.Now().UTC(), time.DateTime))

			ec := utils.LaunchApp("/bin/bash", []string{"-c", fmt.Sprintf("tar cvf - %s | lz4 - %s", dataDirPath, zipFileNameWithoutExt)})
			if ec != 0 {
				utils.ExitWithErrorMsg("ERR: failed to zip data dir")
				return
			}
		},
	}

	return cmd
}
