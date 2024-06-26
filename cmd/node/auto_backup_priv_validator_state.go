package node

import (
	"fmt"
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"time"
)

const (
	flagKeep                   = "keep"
	flagBinaryKillByAutoBackup = "binary"
	flagGenSetup               = "gen-setup"
)

const (
	commandAutoBackupPrivValidatorState = "auto-backup-priv-validator-state-json"
)

const (
	backupPrivValStateJsonPrefixFileName = "priv_validator_state"
	latestBackupPrivValStateJsonFileName = backupPrivValStateJsonPrefixFileName + "_latest.json"
)

func GetAutoBackupPrivValidatorStateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     commandAutoBackupPrivValidatorState + " [node_home]",
		Aliases: []string{"auto-backup-pvs", "auto-backup-priv-validator-state"},
		Short:   "Designed to be run as a service, it will automatically backup the `priv_validator_state.json`, and kill the node process if the content of the file is decreased",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			nodeHomeDirectory := strings.TrimSuffix(strings.TrimSpace(args[0]), "/")
			validateNodeHomeDirectory(nodeHomeDirectory)
			if !strings.Contains(nodeHomeDirectory, "/") {
				utils.ExitWithErrorMsg("ERR: node home directory must be absolute path, eg: /home/user/.nodeHome")
				return
			}

			currentUser, err := user.Current()
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to get current user:", err)
				return
			}
			userHomeDir := currentUser.HomeDir

			if !strings.HasPrefix(nodeHomeDirectory, userHomeDir) {
				utils.ExitWithErrorMsg("ERR: node home directory must be under user home directory:", userHomeDir)
				return
			}

			backupDstPath := path.Join(userHomeDir, fmt.Sprintf(".backup_priv_validator_state_%s", constants.BINARY_NAME))
			createBackupDirIfNotExists(backupDstPath)
			fmt.Println("INF: backup directory:", backupDstPath)

			keepRecent, _ := cmd.Flags().GetInt(flagKeep)
			if keepRecent < 3 {
				keepRecent = 3
			}
			fmt.Println("INF: keep backup of the last", keepRecent, "blocks")

			binaryPathToKill, _ := cmd.Flags().GetString(flagBinaryKillByAutoBackup)
			if binaryPathToKill == "" {
				utils.ExitWithErrorMsg("ERR: required flag --" + flagBinaryKillByAutoBackup)
				return
			}
			if !strings.Contains(binaryPathToKill, "/") {
				utils.ExitWithErrorMsg("ERR: binary name must be absolute path, eg: /home/user/go/bin/" + binaryPathToKill)
				return
			}
			_, exists, isDir, err := utils.FileInfo(binaryPathToKill)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check binary file:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: binary file does not exist:", binaryPathToKill)
				return
			}
			if isDir {
				utils.ExitWithErrorMsg("ERR: specify binary path is a directory:", binaryPathToKill)
				return
			}
			_, binaryNameToKill := path.Split(binaryPathToKill)
			fmt.Println("INF: binary to kill:", binaryNameToKill, "at", binaryPathToKill)

			if cmd.Flags().Changed(flagGenSetup) {
				genSetupThenExit(nodeHomeDirectory, binaryPathToKill, keepRecent, currentUser)
				return
			}

			privValStateJsonFilePath := path.Join(nodeHomeDirectory, "data", "priv_validator_state.json")
			fmt.Println("INF: priv_validator_state.json file path:", privValStateJsonFilePath)

			latestBackupPvs := loadLatestBackupPrivValidatorStateOrExitWithErr(backupDstPath)
			fmt.Println("INF: latest state from backup:")
			fmt.Println(latestBackupPvs.Json())

			const interval = 200 * time.Millisecond
			var lastExecution time.Time

			type backupPerHeight struct {
				heightStr string
				files     []string
			}
			backupFilesByHeight := make([]*backupPerHeight, 0)

			for {
				if time.Since(lastExecution) < interval {
					time.Sleep(20 * time.Millisecond)
					continue
				}

				lastExecution = time.Now().UTC()

				// Remove old backups
				if numberOfBackupHeights := len(backupFilesByHeight); numberOfBackupHeights > keepRecent {
					pruneSize := numberOfBackupHeights - keepRecent
					for _, backupPerHeight := range backupFilesByHeight[:pruneSize] {
						for _, file := range backupPerHeight.files {
							if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
								utils.PrintlnStdErr("ERR: failed to remove backup file", file, ":", err)
								// ignore error
							}
						}
					}
					backupFilesByHeight = backupFilesByHeight[pruneSize:]
				}

				// Load the recent state

				loadRecentPrivateValidatorState := func() (types.PrivateValidatorState, error) {
					pvs := &types.PrivateValidatorState{}
					err := pvs.LoadFromJSONFile(privValStateJsonFilePath)
					if err != nil {
						return types.PrivateValidatorState{}, err
					}
					return *pvs, nil
				}

				recentPvs, err := loadRecentPrivateValidatorState()
				if err != nil {
					utils.PrintlnStdErr("ERR: failed to load priv_validator_state.json file:", err)
					continue
				}

				cmp, _ := latestBackupPvs.CompareState(recentPvs)
				// TODO handle different signs flag, returned by CompareState

				stateNotChanged := cmp == 0 // latest equal to recent
				stateIncreased := cmp < 0   // latest less than recent
				// stateDecreased := cmp > 0   // latest greater than recent

				if stateNotChanged {
					// nothing changed
					continue
				}

				// backup the recent state to file, marked by time and height/round/step
				backupFileNameMarkByTimeAndHrs := fmt.Sprintf(
					"%s_%s_hrs_%s_%d_%d.json",
					backupPrivValStateJsonPrefixFileName,
					utils.GetDateTimeStringCompatibleWithFileName(time.Now().UTC(), time.DateTime),
					recentPvs.Height, recentPvs.Round, recentPvs.Step,
				)
				backupMarkByTimeAndHrsFilePath := path.Join(backupDstPath, backupFileNameMarkByTimeAndHrs)
				err = recentPvs.SaveToJSONFile(backupMarkByTimeAndHrsFilePath)
				if err != nil {
					utils.PrintlnStdErr("ERR: failed to save backup file", backupMarkByTimeAndHrsFilePath, err)
				} else if size := len(backupFilesByHeight); size == 0 || backupFilesByHeight[size-1].heightStr != recentPvs.Height {
					backupFilesByHeight = append(backupFilesByHeight, &backupPerHeight{
						heightStr: recentPvs.Height,
						files:     []string{backupMarkByTimeAndHrsFilePath},
					})
				} else {
					backupFilesByHeight[size-1].files = append(backupFilesByHeight[size-1].files, backupMarkByTimeAndHrsFilePath)
				}

				if stateIncreased {
					// backup the recent state to file, marked by latest
					backupLatestFilePath := path.Join(backupDstPath, latestBackupPrivValStateJsonFileName)
					err = recentPvs.SaveToJSONFile(backupLatestFilePath)
					if err != nil {
						utils.PrintlnStdErr("ERR: failed to save backup file", backupLatestFilePath, err)
					}

					latestBackupPvs = recentPvs
					continue
				}

				// state decreased

				const slightlySleepDuration = 5 * time.Millisecond // prevent consuming all CPU

				if recentPvs.IsEmpty() {
					// mode restore snapshot

					fmt.Println("WARN: detected state file is empty, possibly restoring snapshot")
					fmt.Println("WARN: attempts to kill the node binary", binaryNameToKill, "while waiting content to be restored")

					// possibly restoring snapshot progress
					killedStatusOnSoftProtectRestoreSnapshot := &killedStatus{}

					for {
						shouldIgnoreSleep := killNodeOnLoop(binaryNameToKill, false, killedStatusOnSoftProtectRestoreSnapshot)
						if shouldIgnoreSleep {
							time.Sleep(slightlySleepDuration)
						} else {
							time.Sleep(100 * time.Millisecond)
						}

						recentPvs, err = loadRecentPrivateValidatorState()
						if err != nil {
							utils.PrintlnStdErr("ERR: failed to load priv_validator_state.json file after killing node:", err)
							time.Sleep(slightlySleepDuration)
							continue
						}

						if !recentPvs.IsEmpty() {
							// recent state no longer empty, continue to check in next loop
							break
						}
					}

					lastExecution = time.Time{} // reset last execution time, move to next as fast as possible
					continue
				}

				// mode fatal

				utils.PrintlnStdErr("FATAL: priv_validator_state.json content decreased")
				utils.PrintlnStdErr("Previous state:")
				utils.PrintlnStdErr(latestBackupPvs.Json())
				utils.PrintlnStdErr("Recent state:")
				utils.PrintlnStdErr(recentPvs.Json())

				go func(latestBackupPvs, recentPvs types.PrivateValidatorState) {
					// launch another goroutine to go to kill process as fast as possible
					reportMismatchFilePath := path.Join(backupDstPath, fmt.Sprintf("mismatch_%s_%s.json", backupPrivValStateJsonPrefixFileName, utils.GetDateTimeStringCompatibleWithFileName(time.Now().UTC(), time.DateTime)))
					content := fmt.Sprintf(`Previous state:
%s

Recent state:
%s
`, latestBackupPvs.Json(), recentPvs.Json())
					for {
						// write the mismatch content to file
						err := os.WriteFile(reportMismatchFilePath, []byte(content), 0o644)
						if err != nil {
							utils.PrintlnStdErr("ERR: failed to write mismatch file:", reportMismatchFilePath, ":", err)
							time.Sleep(300 * time.Millisecond)
							continue
						}
						break
					}
					fmt.Println("INF: mismatch content written to file:", reportMismatchFilePath)
				}(latestBackupPvs, recentPvs)

				go func(latestBackupPvs, recentPvs types.PrivateValidatorState, userHomeDir string) {
					// launch another goroutine to go to kill process as fast as possible
					urgentReportMismatchFilePath := path.Join(userHomeDir, "FATAL_REPORT_MISMATCH_PRIV_VALIDATOR_STATE.txt")
					content := fmt.Sprintf(`
%s detected a mismatch in priv_validator_state.json content, currently executing killing the node!!!

Previous state:
%s

Recent state:
%s

How to recover:
- Fix your problem in priv_validator_state.json, can check latest backup at %s
- Stop this auto-backup service
- Restart the node
- Restart this auto-backup service
`, constants.BINARY_NAME, latestBackupPvs.Json(), recentPvs.Json(), backupDstPath)
					for {
						// write the report to file
						err := os.WriteFile(urgentReportMismatchFilePath, []byte(content), 0o644)
						if err != nil {
							utils.PrintlnStdErr("ERR: failed to write report file:", urgentReportMismatchFilePath, ":", err)
							time.Sleep(300 * time.Millisecond)
							continue
						}
						break
					}
					fmt.Println("INF: report content written to file:", urgentReportMismatchFilePath)
				}(latestBackupPvs, recentPvs, userHomeDir)

				// Force-stop the node

				killedStatusOnFatal := &killedStatus{}
				fmt.Println("WARN: Killing the node binary:", binaryNameToKill)
				for {
					shouldIgnoreSleep := killNodeOnLoop(binaryNameToKill, true, killedStatusOnFatal)
					if shouldIgnoreSleep {
						time.Sleep(slightlySleepDuration)
					} else {
						time.Sleep(300 * time.Millisecond)
					}
				}
			}
		},
	}

	cmd.Flags().Int(flagKeep, 3, "Keep backup of the last N blocks")
	cmd.Flags().String(flagBinaryKillByAutoBackup, "", "Absolute path of the chain binary to be killed by process when priv_validator_state.json has problem")
	cmd.Flags().Bool(flagGenSetup, false, "Display guide to setup instead of running business logic")

	return cmd
}

type killedStatus struct {
	killedCount uint
}

func killNodeOnLoop(binaryNameToKill string, fatalCase bool, killedStatus *killedStatus) (shouldIgnoreSleep bool) {
	processes, err := process.Processes()
	if err != nil {
		utils.PrintlnStdErr("ERR: failed to get processes:", err)
		shouldIgnoreSleep = true
		return
	}

	var processesToKill []*process.Process
	for _, p := range processes {
		func(p *process.Process) {
			var includeToBeKilled bool
			defer func() {
				if !includeToBeKilled {
					return
				}
				processesToKill = append(processesToKill, p)
			}()

			var sameName, hasStart bool

			cmdLine, _ := p.Cmdline()
			name, _ := p.Name()

			if strings.Contains(cmdLine, binaryNameToKill) && strings.Contains(cmdLine, " start") {
				sameName = true
				hasStart = true
			}

			if !sameName && name == binaryNameToKill {
				sameName = true
			}

			if !hasStart && strings.Contains(cmdLine, " start") {
				hasStart = true
			}

			if !sameName || !hasStart {
				cmdLineSlice, _ := p.CmdlineSlice()
				for i := 0; i < len(cmdLineSlice); i++ {
					arg := cmdLineSlice[i]
					if strings.Contains(arg, binaryNameToKill) && strings.Contains(arg, " start") {
						sameName = true
						hasStart = true
					}

					if !sameName && strings.Contains(arg, binaryNameToKill) {
						sameName = true
					}

					if !hasStart && strings.Contains(arg, "start") {
						hasStart = true
					}

					if sameName && hasStart {
						break
					}
				}
			}

			includeToBeKilled = sameName && hasStart
		}(p)
	}

	if len(processesToKill) < 1 {
		if fatalCase && killedStatus.killedCount < 1 {
			utils.PrintlnStdErr("ERR: no process found to be killed")
		}
		shouldIgnoreSleep = true
		return
	}

	var sbKill strings.Builder
	for i, p := range processesToKill {
		if i > 0 {
			sbKill.WriteString(" ; ")
		}
		sbKill.WriteString("kill -9 ")
		sbKill.WriteString(fmt.Sprintf("%d", p.Pid))
	}

	var anyError bool

	cmd := exec.Command("/bin/bash", "-c", sbKill.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println(sbKill.String())
	errKill9 := cmd.Start()
	if errKill9 != nil {
		anyError = true
		utils.PrintlnStdErr("ERR: failed to start command kill -9")
	}

	for _, p := range processesToKill {
		fmt.Println("WARN: killing", p.Pid)
		errLibKill := p.Kill()
		if errLibKill != nil {
			anyError = true
			utils.PrintlnStdErr("ERR: failed to kill", p.Pid, ":", err)
		}
	}

	if errKill9 == nil {
		_ = cmd.Wait()
	}

	if !anyError {
		killedStatus.killedCount += uint(len(processesToKill))
	}

	fmt.Println("INF: total killed", killedStatus.killedCount, "processes")

	return
}

func createBackupDirIfNotExists(backupDstPath string) {
	_, exists, isDir, err := utils.FileInfo(backupDstPath)
	if err != nil {
		utils.ExitWithErrorMsg("ERR: failed to check backup directory:", err)
		return
	}
	if !exists {
		err = os.Mkdir(backupDstPath, 0o700)
		if err != nil {
			utils.ExitWithErrorMsg("ERR: failed to create backup directory at", backupDstPath, ":", err)
			return
		}
		_, exists, isDir, err = utils.FileInfo(backupDstPath)
		if err != nil {
			utils.ExitWithErrorMsg("ERR: failed to check backup directory after created:", err)
			return
		}
		if !exists {
			utils.ExitWithErrorMsg("ERR: backup directory does not exists after create:", backupDstPath)
			return
		}
	}
	if !isDir {
		utils.ExitWithErrorMsg("ERR: backup directory is not a directory:", backupDstPath)
		return
	}
}

func loadLatestBackupPrivValidatorStateOrExitWithErr(backupDstPath string) types.PrivateValidatorState {
	filePath := path.Join(backupDstPath, latestBackupPrivValStateJsonFileName)
	_, exists, _, err := utils.FileInfo(filePath)
	if err != nil {
		utils.ExitWithErrorMsg("ERR: failed to check latest backup file", latestBackupPrivValStateJsonFileName, ":", err)
		return types.PrivateValidatorState{}
	}
	if !exists {
		return types.NewEmptyPrivateValidatorState()
	}

	pvs := &types.PrivateValidatorState{}
	err = pvs.LoadFromJSONFile(filePath)
	if err != nil {
		utils.ExitWithErrorMsg("ERR: failed to load latest backup file", latestBackupPrivValStateJsonFileName, ":", err)
		return types.PrivateValidatorState{}
	}

	return *pvs
}

func genSetupThenExit(nodeHomeDirectory, binaryPathToKill string, keepRecent int, currentUser *user.User) {
	const serviceFileName = "auto-backup-pvs"
	fmt.Println("Input chain name (eg: Cosmos Hub):")
	chainName := utils.ReadText(false)
	fmt.Println("Mainnet or Testnet?")
	networkType := utils.ReadText(false)
	fmt.Println()
	fmt.Println("INF: setup guide:")
	fmt.Println()
	fmt.Println("1. Create service file")
	fmt.Println("> sudo vi /etc/systemd/system/" + serviceFileName + ".service")
	fmt.Printf(`[Unit]
Description=Auto backup priv_validator_state.json for Validator on %s %s
After=network.target
#
[Service]
User=%s
ExecStart=/usr/local/bin/%s node %s %s --%s %s --%s %d
RestartSec=1
Restart=on-failure
LimitNOFILE=1024
#
[Install]
WantedBy=multi-user.target
`, chainName, networkType, currentUser.Username, constants.BINARY_NAME, commandAutoBackupPrivValidatorState, nodeHomeDirectory, flagBinaryKillByAutoBackup, binaryPathToKill, flagKeep, keepRecent)
	fmt.Println()
	fmt.Println("2. Setup visudo")
	fmt.Println()
	fmt.Println("> sudo visudo")
	fmt.Printf(strings.ReplaceAll(strings.ReplaceAll(`# Allow user @USER@ to manage @SVC@ service
@USER@ ALL= NOPASSWD: /usr/bin/systemctl start @SVC@
@USER@ ALL= NOPASSWD: /usr/bin/systemctl stop @SVC@
@USER@ ALL= NOPASSWD: /usr/bin/systemctl restart @SVC@
@USER@ ALL= NOPASSWD: /usr/bin/systemctl enable @SVC@
# Do not allow disable
@USER@ ALL= NOPASSWD: /usr/bin/systemctl status @SVC@
`, "@USER@", currentUser.Username), "@SVC@", serviceFileName))
	fmt.Println()
	fmt.Println("3. Enable service to automatically run at startup")
	fmt.Println()
	fmt.Println("> sudo systemctl daemon-reload && sudo systemctl enable " + serviceFileName)
	os.Exit(0)
}
