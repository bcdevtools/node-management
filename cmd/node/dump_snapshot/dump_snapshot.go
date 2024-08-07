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
	flagFixGenesis                  = "fix-genesis"
)

func GetDumpSnapshotCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dump-snapshot [node_home]",
		Short: "Dump snapshot from node, using Cosmos-SDK snapshot commands",
		Long: "Dump snapshot from node, using Cosmos-SDK snapshot commands.\n" +
			"The node will be stopped, data will be exported, dumped and restore into another node home directory (eg ~/.gaia → ~/.gaia-dump).",
		Args: cobra.ExactArgs(1),
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
				utils.PrintfStdErr("ERR: correct flag --%s\n", flagBinary)
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

			const minOfMaxDuration = 30 * time.Minute
			if maxDuration < minOfMaxDuration {
				utils.PrintfStdErr("ERR: minimum accepted for --%s is %s\n", flagMaxDuration, minOfMaxDuration)
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

			var stoppedService bool

			defer func() {
				if exitWithError && stoppedService {
					fmt.Println("INF: restarting service before exit due to error")
					ec := utils.LaunchApp("sudo", []string{"systemctl", "restart", serviceName})
					if ec != 0 {
						utils.PrintlnStdErr("ERR: failed to restart service", serviceName)
					}
				}
			}()

			go func() {
				time.Sleep(maxDuration)
				utils.PrintlnStdErr("ERR: timeout")
				execCleanup()
				if stoppedService {
					fmt.Println("INF: restarting service before exit due to timeout")
					ec := utils.LaunchApp("sudo", []string{"systemctl", "restart", serviceName})
					if ec != 0 {
						utils.PrintlnStdErr("ERR: failed to restart service", serviceName)
					}
				}
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
			if cmd.Flags().Changed(flagFixGenesis) {
				fmt.Println("INF: fixing genesis initial_height")
				genesisFilePath := path.Join(dumpHomeDir, "config", "genesis.json")
				_ = utils.LaunchApp("/bin/bash", []string{
					"-c", fmt.Sprintf(`jq '.initial_height = "1"' %s > %s.tmp && mv %s.tmp %s`, genesisFilePath, genesisFilePath, genesisFilePath, genesisFilePath),
				})
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
				stoppedService = true
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

			if !noService {
				// restart the service
				fmt.Println("INF: restarting service")
				ec = utils.LaunchApp("sudo", []string{"systemctl", "restart", serviceName})
				if ec != 0 {
					utils.PrintlnStdErr("ERR: failed to restart service")
					exitWithError = true
					return
				}
				time.Sleep(15 * time.Second)
			}

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
			}

			rpc, err := types.ReadNodeRpcFromConfigToml(path.Join(nodeHomeDirectory, "config", "config.toml"))
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to read node rpc from config.toml:", err)
				exitWithError = true
				return
			}
			externalRPCs, _ := cmd.Flags().GetStringSlice(flagExternalRpc)
			rpcEps := []string{rpc}
			for _, externalRPC := range externalRPCs {
				rpc := strings.TrimSpace(externalRPC)
				if rpc == "" {
					continue
				}
				rpcEps = append(rpcEps, rpc)
			}

			for {
				resp, err := http.Get(fmt.Sprintf("%s/status", strings.TrimSuffix(rpc, "/")))
				if err == nil && resp.StatusCode == http.StatusOK {
					break
				}
				fmt.Println("INF: waiting node up")
				time.Sleep(10 * time.Second)
			}

			chanTrustHash := make(chan string, len(rpcEps))
			registeredCleanup = append(registeredCleanup, func() {
				close(chanTrustHash)
			})

			for _, rpc := range rpcEps {
				go func(rpc string) {
					var trustHash string
					defer func() {
						chanTrustHash <- trustHash
					}()

					output, ec := utils.LaunchAppAndGetOutput(
						"/bin/bash",
						[]string{
							"-c", fmt.Sprintf(
								`curl -m 30 -s "%s/block?height=%d" | jq -r .result.block_id.hash`,
								rpc,
								mostRecentSnapshot.height,
							),
						},
					)
					if ec != 0 {
						utils.PrintlnStdErr(output)
						utils.PrintlnStdErr("ERR: failed to get block hash from rpc:", rpc)
						return
					}

					trustHash = strings.TrimSpace(output)
					if !regexp.MustCompile(`^[A-F\d]{64}$`).MatchString(trustHash) {
						utils.PrintlnStdErr("ERR: invalid block hash", trustHash, "from rpc", rpc)
						trustHash = ""
					}
				}(rpc)
			}

			var trustHash string
			for c := 1; c <= len(rpcEps); c++ {
				resTrustHash := <-chanTrustHash
				if resTrustHash == "" {
					continue
				}
				if trustHash != "" {
					// take the first valid trust hash
					continue
				}
				trustHash = resTrustHash
			}

			if trustHash == "" {
				utils.PrintlnStdErr("ERR: failed to get trust hash from rpc")
				exitWithError = true
				return
			}

			if len(rpcEps) == 1 {
				rpcEps = append(rpcEps, rpcEps[0])
			}
			sedArgs := []string{
				"-i.bak",
				"-E", `s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"` + strings.Join(func() []string {
					if len(rpcEps) == 1 {
						return []string{rpcEps[0], rpcEps[0]}
					}
					return rpcEps
				}(), ",") + `\"| ; s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1` + mostRecentSnapshot.HeightStr() + `| ; s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1"` + trustHash + `"|`,
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

			fmt.Println("INF: successfully dumped snapshot into", dumpHomeDir)
		},
	}

	cmd.Flags().String(flagBinary, "", "Path to the binary")
	cmd.Flags().Duration(flagMaxDuration, 1*time.Hour, "Maximum duration to wait for dumping snapshot")
	cmd.Flags().Bool(flagNoService, false, "Do not stop and start service")
	cmd.Flags().String(flagServiceName, "", "Custom service name, used to call start/stop")
	cmd.Flags().StringSlice(flagExternalRpc, []string{}, "External RPC address used for bootstrapping node")
	cmd.Flags().Bool(flagXCrisisSkipAssertInvariants, false, "Skip assert invariants")
	cmd.Flags().Bool(flagFixGenesis, false, "Fix `initial_height` in genesis.json")

	return cmd
}
