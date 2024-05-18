package node

import (
	"fmt"
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"time"
)

const (
	flagBinary                  = "binary"
	flagBackupPrivValStateJson  = "backup-pvs"
	flagRestorePrivValStateJson = "restore-pvs"
)

const fileNamePrivValState = "priv_validator_state.json"

func GetPruneNodeDataCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "prune-data [node_home]",
		Short: "Prune node data.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			inputHomeDirectory := strings.TrimSpace(args[0])
			binary, _ := cmd.Flags().GetString(flagBinary)
			backupPrivValStateJson, _ := cmd.Flags().GetString(flagBackupPrivValStateJson)
			restorePrivValStateJson, _ := cmd.Flags().GetBool(flagRestorePrivValStateJson)

			if inputHomeDirectory == "" {
				utils.ExitWithErrorMsg("ERR: required input home directory")
				return
			}

			if binary == "" {
				utils.ExitWithErrorMsgf("ERR: required flag --%s\n", flagBinary)
				return
			}
			if strings.Contains(binary, "/") {
				_, exists, isDir, err := utils.FileInfo(binary)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to check binary path:", err)
					return
				}
				if !exists {
					utils.ExitWithErrorMsg("ERR: specified binary does not exist:", binary)
					return
				}
				if isDir {
					utils.ExitWithErrorMsg("ERR: specified binary is a directory:", binary)
					return
				}
			} else if !utils.HasBinaryName(binary) {
				utils.ExitWithErrorMsg("ERR: binary name ", binary, "might not available in $PATH")
			}

			ensureDirExists := func(path string) {
				if _, exists, isDir, err := utils.FileInfo(path); err != nil {
					utils.ExitWithErrorMsg("ERR: failed to check directory:", err)
					return
				} else if !exists {
					utils.ExitWithErrorMsg("ERR: required directory does not exist:", path)
					return
				} else if !isDir {
					utils.ExitWithErrorMsg("ERR: specified path is not a directory:", path)
					return
				}
			}

			configDir := path.Join(inputHomeDirectory, "config")
			dataDir := path.Join(inputHomeDirectory, "data")

			ensureDirExists(configDir)
			ensureDirExists(dataDir)

			filePathPrivValState := path.Join(dataDir, fileNamePrivValState)
			if _, exists, isDir, err := utils.FileInfo(filePathPrivValState); err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check", filePathPrivValState, ":", err)
				return
			} else if !exists || isDir {
				utils.ExitWithErrorMsg("ERR:", filePathPrivValState, "is not exists or not a file")
				return
			}

			bz, err := os.ReadFile(filePathPrivValState)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to read file", filePathPrivValState, ":", err)
				return
			}
			if len(bz) < 1 {
				utils.ExitWithErrorMsg("ERR: file is empty:", filePathPrivValState)
				return
			}

			pvs := &types.PrivateValidatorState{}
			err = pvs.LoadFromJSONFile(filePathPrivValState)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to load", filePathPrivValState, ": %v\n", err)
				return
			}

			var additionalBackupPrivStateJsonFilePath string

			if pvs.IsEmpty() {
				fmt.Println("INF:", fileNamePrivValState, "is empty, location:", filePathPrivValState)
				if restorePrivValStateJson {
					utils.ExitWithErrorMsgf("ERR: %s is empty, usage of flag --%s is prohibited\n", fileNamePrivValState, flagRestorePrivValStateJson)
				}
			} else {
				fmt.Println("INF:", fileNamePrivValState, "is not empty, location:", filePathPrivValState)
				fmt.Println(string(bz))
				if backupPrivValStateJson == "" {
					utils.ExitWithErrorMsgf("ERR: require backup file via flag --%s, to prove you already backup the file %s\n", flagBackupPrivValStateJson, fileNamePrivValState)
					return
				}

				backupPrivValStateJsonDir, _ := path.Split(backupPrivValStateJson)
				backupPrivValStateJsonDir = strings.TrimSuffix(backupPrivValStateJsonDir, "/")
				if backupPrivValStateJsonDir == strings.TrimSuffix(dataDir, "/") {
					utils.ExitWithErrorMsg("ERR: backup file must not be in data directory")
					return
				}

				bzOfBackup, err := os.ReadFile(backupPrivValStateJson)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to read backup file", backupPrivValStateJson, ":", err)
					return
				}
				if len(bzOfBackup) < 1 {
					utils.ExitWithErrorMsg("ERR: backup file is empty:", backupPrivValStateJson)
					return
				}

				pvsBackup := &types.PrivateValidatorState{}
				err = pvsBackup.LoadFromJSONFile(backupPrivValStateJson)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to load backup file", backupPrivValStateJson, ":", err)
					return
				}

				if pvsBackup.IsEmpty() {
					utils.ExitWithErrorMsg("ERR: backup file is empty", backupPrivValStateJson)
					return
				}

				if !pvs.Equals(pvsBackup) {
					utils.ExitWithErrorMsg("ERR: backup file at", backupPrivValStateJson, "has different content with", filePathPrivValState)
					return
				}

				fmt.Println("INF:", filePathPrivValState)
				fmt.Println(string(bz))
				fmt.Println("INF:", backupPrivValStateJson)
				fmt.Println(string(bzOfBackup))

				// create additional backup
				additionalBackupFile := path.Join(configDir, fmt.Sprintf("%s.%s.%s.bak", fileNamePrivValState, utils.GetDateTimeStringCompatibleWithFileName(time.Now().UTC(), time.DateTime), constants.BINARY_NAME))

				fmt.Println("Going to create an additional backup of", fileNamePrivValState)
				fmt.Println("at", additionalBackupFile)

				err = os.WriteFile(additionalBackupFile, bz, 0644)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to write additional backup file:", err)
					return
				}

				backupPsv2 := &types.PrivateValidatorState{}
				err = backupPsv2.LoadFromJSONFile(additionalBackupFile)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to load additional backup file", additionalBackupFile, ":", err)
					return
				}
				if !pvs.Equals(backupPsv2) {
					utils.ExitWithErrorMsg("ERR: additional backup file at", additionalBackupFile, "has different content with", filePathPrivValState)
					return
				}

				fmt.Println("Additional backup file created successfully!")

				additionalBackupPrivStateJsonFilePath = additionalBackupFile
			}

			pruneArgs := []string{"tendermint", "unsafe-reset-all", "--home", inputHomeDirectory, "--keep-addr-book"}

			const sleepTime = 30 * time.Second
			fmt.Println("INF: Going to run the following command after", sleepTime)
			fmt.Println(">", strings.Join(append([]string{binary}, pruneArgs...), " "))
			if additionalBackupPrivStateJsonFilePath != "" {
				fmt.Print("Restore ", fileNamePrivValState, ": ")
				if restorePrivValStateJson {
					fmt.Println("YES")
				} else {
					fmt.Println("NO")
				}
			}
			time.Sleep(sleepTime)

			ec := utils.LaunchApp(binary, pruneArgs)
			if ec != 0 {
				utils.ExitWithErrorMsgf("ERR: %s exited with code %d\n", binary, ec)
				return
			}

			if additionalBackupPrivStateJsonFilePath != "" && restorePrivValStateJson {
				fmt.Println("INF: Restoring", fileNamePrivValState, "from backup file:", additionalBackupPrivStateJsonFilePath)

				pvs := &types.PrivateValidatorState{}
				err = pvs.LoadFromJSONFile(filePathPrivValState)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to load", fileNamePrivValState, ": %v\n", err)
					return
				}
				if !pvs.IsEmpty() {
					utils.ExitWithErrorMsg("ERR:", fileNamePrivValState, "is not empty after pruned data")
					return
				}

				bz, err := os.ReadFile(additionalBackupPrivStateJsonFilePath)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to read backup file", additionalBackupPrivStateJsonFilePath, ":", err)
					return
				}

				pvsBackup := &types.PrivateValidatorState{}
				err = pvsBackup.LoadFromJSONFile(additionalBackupPrivStateJsonFilePath)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to load backup file", additionalBackupPrivStateJsonFilePath, ":", err)
					return
				}
				if pvsBackup.IsEmpty() {
					utils.ExitWithErrorMsg("ERR: backup file is empty", additionalBackupPrivStateJsonFilePath)
					return
				}

				err = os.WriteFile(filePathPrivValState, bz, 0o644)
				if err != nil {
					utils.PrintlnStdErr("ERR: failed to write file", fileNamePrivValState, ":", err)
					utils.PrintlnStdErr(string(bz))
					os.Exit(1)
					return
				}

				fmt.Println("INF: Restored", fileNamePrivValState, "from backup file:", additionalBackupPrivStateJsonFilePath)
				bz, err = os.ReadFile(filePathPrivValState)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to read file", fileNamePrivValState, " for confirmation:", err)
					return
				}
				fmt.Println(string(bz))
			}
		},
	}

	cmd.Flags().String(flagBinary, "", "Binary name")
	cmd.Flags().String(flagBackupPrivValStateJson, "", "Backup "+fileNamePrivValState+" file path to prove you already backup it, required if the file in data is not empty")
	cmd.Flags().Bool(flagRestorePrivValStateJson, false, "Restore "+fileNamePrivValState+" file path after pruned data")

	return cmd
}
