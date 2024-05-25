package dump_snapshot

import (
	"fmt"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	flagNoService                   = "no-service"
	flagServiceName                 = "service-name"
	flagExternalRpc                 = "external-rpc"
	flagBinary                      = "binary"
	flagMaxDuration                 = "max-duration"
	flagXCrisisSkipAssertInvariants = "x-crisis-skip-assert-invariants"
)

func GetDumpSnapshotCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dump-snapshot [node_home]",
		Short: "Dump snapshot from node, using Cosmos-SDK snapshot commands",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			var exitWithError bool
			defer func() {
				if exitWithError {
					os.Exit(1)
				}
			}()

			// Process:
			// - Stop service
			// - Export snapshot
			// - Dump snapshot
			// - Restore snapshot
			// - Start service

			/**
			Coding convention:
			due to resource cleanup, the code is written in a way that ensure the cleanup methods are called,
			so that the resources are released properly.
			Usage of os.Exit() is avoided, and the exitWithError flag is used to indicate the error.
			Usage of functions that call os.Exit() is prohibited.
			*/

			nodeHomeDirectory := strings.TrimSpace(args[0])
			binary, _ := cmd.Flags().GetString(flagBinary)
			maxDuration, _ := cmd.Flags().GetDuration(flagMaxDuration)
			noService, _ := cmd.Flags().GetBool(flagNoService)
			if !noService && !utils.IsLinux() {
				noService = true
				fmt.Printf("INF: --%s is forced on Non-Linux\n", flagNoService)
			}

			if err := validateNodeHomeDirectory(nodeHomeDirectory); err != nil {
				utils.PrintlnStdErr("ERR: invalid node home directory:", err)
				exitWithError = true
				return
			}
			nodeHomeDirectory, err := filepath.Abs(nodeHomeDirectory)
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to get absolute path of node home directory:", err)
				exitWithError = true
				return
			}

			if err := validateBinary(binary); err != nil {
				utils.PrintlnStdErr("ERR: invalid binary path:", err)
				utils.PrintlnStdErr("ERR: correct flag --%s\n", flagBinary)
				exitWithError = true
				return
			}

			serviceName, err := getServiceName(noService, binary, cmd)
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to get service name")
				utils.PrintlnStdErr("ERR:", err.Error())
				exitWithError = true
				return
			}

			if maxDuration < 30*time.Minute {
				utils.PrintlnStdErr("ERR: minimum duration is 30 minutes")
				exitWithError = true
				return
			}

			var registeredCleanup []func()
			execCleanup := func() {
				for i, cleanup := range registeredCleanup {
					func(i int, cleanup func()) {
						defer func() {
							r := recover()
							if err != nil {
								utils.PrintlnStdErr("ERR: panic in cleanup[", i, "]:", r)
							}
						}()

						cleanup()
					}(i, cleanup)
				}
			}

			defer execCleanup()

			go func() {
				time.Sleep(maxDuration)
				utils.PrintlnStdErr("ERR: timeout")
				execCleanup()
				os.Exit(1)
			}()

			parentHomeDir, homeDirName := path.Split(nodeHomeDirectory)

			dumpDirName := homeDirName + "-dump"
			dumpHomeDir := path.Join(parentHomeDir, dumpDirName)

			if err := prepareDumpNodeHomeDirectory(dumpHomeDir, nodeHomeDirectory); err != nil {
				utils.PrintlnStdErr("ERR: failed to prepare dump home directory:", err)
				exitWithError = true
				return
			}

			fmt.Println("INF: force reset dump home directory")
			ec := utils.LaunchApp(
				binary, []string{
					"tendermint", "unsafe-reset-all",
					"--home", dumpHomeDir,
					"--keep-addr-book",
				},
			)
			if ec != 0 {
				utils.PrintlnStdErr("ERR: failed to unsafe-reset-all the dump home directory")
				exitWithError = true
				return
			}

			appOriginalNodeMutex, appDumpNodeMutex, errAcqSi := acquireSingletonInstance(nodeHomeDirectory, dumpHomeDir)
			defer func() {
				if appOriginalNodeMutex != nil {
					appOriginalNodeMutex.ReleaseLockWL()
				}
				if appDumpNodeMutex != nil {
					appDumpNodeMutex.ReleaseLockWL()
				}
			}()
			if errAcqSi != nil {
				utils.PrintlnStdErr("ERR: failed to acquire singleton instance:", errAcqSi)
				exitWithError = true
				return
			}

			if !noService {
				fmt.Println("INF: stopping service")
				ec := utils.LaunchApp("sudo", []string{"systemctl", "stop", serviceName})
				if ec != 0 {
					utils.PrintlnStdErr("ERR: failed to stop service")
					exitWithError = true
					return
				}
				time.Sleep(15 * time.Second) // wait completely shutdown
			}

			fmt.Println("INF: exporting snapshot")
			_ = utils.LaunchApp(
				binary, []string{
					"snapshots", "export", "--home", nodeHomeDirectory,
				},
			)

			fmt.Println("INF: checking snapshots")
			snapshots, err := loadSnapshotList(binary, nodeHomeDirectory)
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to get list after exported snapshot:", err)
				exitWithError = true
				return
			} else if len(snapshots) == 0 {
				utils.PrintlnStdErr("ERR: failed to get list after exported snapshot")
				exitWithError = true
				return
			}

			snapshots.Sort()
			mostRecentSnapshot := snapshots[0]
			fmt.Println("INF: most recent snapshot:", mostRecentSnapshot.height, ", format", mostRecentSnapshot.format, ", chunks", mostRecentSnapshot.chunks)

			outputFileName := fmt.Sprintf("dump-snapshot.%d-%d.tar.gz", mostRecentSnapshot.height, mostRecentSnapshot.format)
			fmt.Println("INF: dumping snapshot")
			ec = utils.LaunchApp(binary, []string{
				"snapshots", "dump", mostRecentSnapshot.HeightStr(), mostRecentSnapshot.FormatStr(),
				"--home", nodeHomeDirectory,
				"--output", outputFileName,
			})
			if ec != 0 {
				utils.PrintlnStdErr("ERR: failed to dump snapshot")
				exitWithError = true
				return
			}
			if err := validateOutputFile(outputFileName); err != nil {
				utils.PrintlnStdErr("ERR: failed to validate output file:", err)
				exitWithError = true
				return
			}

			fmt.Println("INF: snapshot dumped successfully:", outputFileName)

			registeredCleanup = append(registeredCleanup, func() {
				_ = os.Remove(outputFileName)
			})

			fmt.Println("INF: restoring into", dumpHomeDir)
			ec = utils.LaunchApp(binary, []string{
				"snapshots", "load", outputFileName,
				"--home", dumpHomeDir,
			})
			if ec != 0 {
				utils.PrintlnStdErr("ERR: failed to load snapshot")
				exitWithError = true
				return
			}

			snapshots, err = loadSnapshotList(binary, dumpHomeDir)
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to get list after loaded snapshot:", err)
				exitWithError = true
				return
			} else if len(snapshots) == 0 {
				utils.PrintlnStdErr("ERR: failed to get list after loaded snapshot")
				exitWithError = true
				return
			}

			ec = utils.LaunchApp(binary, []string{
				"snapshots", "restore", mostRecentSnapshot.HeightStr(), mostRecentSnapshot.FormatStr(),
				"--home", dumpHomeDir,
			})
			if ec != 0 {
				utils.PrintlnStdErr("ERR: failed to restore snapshot")
				exitWithError = true
				return
			}

			if noService {
				// launching node for bootstrapping
				startArgs := []string{
					"start",
					"--home", nodeHomeDirectory,
				}
				if cmd.Flags().Changed(flagXCrisisSkipAssertInvariants) {
					startArgs = append(startArgs, fmt.Sprintf("--%s", flagXCrisisSkipAssertInvariants))
				}
				launchCmd := exec.Command(binary, startArgs...)
				launchCmd.Stdout = os.Stdout
				launchCmd.Stderr = os.Stderr
				err = launchCmd.Start()
				if err != nil {
					utils.PrintlnStdErr("ERR: failed to start node for bootstrapping:", err)
					exitWithError = true
					return
				}
				registeredCleanup = append(registeredCleanup, func() {
					_ = launchCmd.Process.Kill()
				})
			} else {
				fmt.Println("INF: restarting service")
				ec = utils.LaunchApp("sudo", []string{"systemctl", "restart", serviceName})
				if ec != 0 {
					utils.PrintlnStdErr("ERR: failed to restart service")
					exitWithError = true
					return
				}
				time.Sleep(5 * time.Second)
			}

			rpc, err := types.ReadNodeRpcFromConfigToml(path.Join(nodeHomeDirectory, "config", "config.toml"))
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to read node rpc from config.toml:", err)
				exitWithError = true
				return
			}
			externalRpcs, _ := cmd.Flags().GetStringSlice(flagExternalRpc)
			rpcEps := append(externalRpcs, rpc)
			if len(rpcEps) == 1 {
				rpcEps = append(rpcEps, rpcEps[0])
			}

			for {
				resp, err := http.Get(fmt.Sprintf("%s/status", strings.TrimSuffix(rpc, "/")))
				if err == nil && resp.StatusCode == http.StatusOK {
					break
				}
				fmt.Println("INF: waiting node up")
				time.Sleep(10 * time.Second)
			}

			output, ec := utils.LaunchAppAndGetOutput(
				"/bin/bash",
				[]string{
					"-c", fmt.Sprintf(`curl -s "%s/block?height=%d" | jq -r .result.block_id.hash`, rpc, mostRecentSnapshot.height),
				},
			)
			if ec != 0 {
				utils.PrintlnStdErr(output)
				utils.PrintlnStdErr("ERR: failed to get block hash from rpc:", rpc)
				exitWithError = true
				return
			}
			trustHash := strings.TrimSpace(output)
			if !regexp.MustCompile(`^[A-F\d]{64}$`).MatchString(trustHash) {
				utils.PrintlnStdErr("ERR: invalid block hash", trustHash, "from rpc")
				exitWithError = true
				return
			}
			sedArgs := []string{
				"-i.bak",
				"-E", `s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"` + strings.Join(rpcEps, ",") + `\"| ; s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1` + fmt.Sprintf("%d", mostRecentSnapshot.height) + `| ; s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1"` + trustHash + `"|`,
				path.Join(dumpHomeDir, "config", "config.toml"),
			}
			ec = utils.LaunchApp("sed", sedArgs)
			if ec != 0 {
				utils.PrintlnStdErr("ERR: failed to launch sed to update config file")
				utils.PrintlnStdErr("> sed " + strings.Join(sedArgs, " "))
				exitWithError = true
				return
			}

			fmt.Println("INF: bootstrapping snapshot")
			ec = utils.LaunchApp(binary, []string{
				"tendermint", "bootstrap-state",
				"--home", dumpHomeDir,
			})
			if ec != 0 {
				utils.PrintlnStdErr("ERR: failed to bootstrap snapshot")
				exitWithError = true
				return
			}

			fmt.Println("INF: dumped successfully")
		},
	}

	cmd.Flags().String(flagBinary, "", "Path to the binary")
	cmd.Flags().Duration(flagMaxDuration, 12*time.Hour, "Maximum duration to wait for dumping snapshot")
	cmd.Flags().Bool(flagNoService, false, "Do not stop and start service")
	cmd.Flags().String(flagServiceName, "", "Custom service name, used to call start/stop")
	cmd.Flags().StringSlice(flagExternalRpc, []string{}, "External RPC address used for bootstrapping node")
	cmd.Flags().Bool(flagXCrisisSkipAssertInvariants, false, "Skip assert invariants")

	return cmd
}
