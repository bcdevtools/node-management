package node

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	flagAllowLocalBinary = "allow-local-binary"
	flagAddressBook      = "address-book"
	flagPeers            = "peers"
	flagSeeds            = "seeds"
	flagRpc              = "rpc"
	flagMaxDuration      = "max-duration"
)

func GetStateSyncCmd() *cobra.Command {
	msgDescFlagAllowLocalBinary := fmt.Sprintf("binary must be specified with full path, belong to home dir of another user, use flag --%s to allow local binary. This was designed to prevent mis-match binary version between users of same machine", flagAllowLocalBinary)

	var cmd = &cobra.Command{
		Use:   "state-sync [node_home]",
		Short: "Start state sync for a node",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			nodeHomeDirectory := strings.TrimSpace(args[0])
			binary, _ := cmd.Flags().GetString(flagBinary)
			allowLocalBinary, _ := cmd.Flags().GetBool(flagAllowLocalBinary)
			addressBookFilePath, _ := cmd.Flags().GetString(flagAddressBook)
			newPeers, _ := cmd.Flags().GetString(flagPeers)
			seeds, _ := cmd.Flags().GetString(flagSeeds)
			rpc, _ := cmd.Flags().GetString(flagRpc)
			maxDuration, _ := cmd.Flags().GetDuration(flagMaxDuration)

			if nodeHomeDirectory == "" {
				utils.ExitWithErrorMsg("ERR: required input home directory")
				return
			}
			_, exists, isDir, err := utils.FileInfo(nodeHomeDirectory)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check node home directory:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: node home directory does not exist:", nodeHomeDirectory)
				return
			}
			if !isDir {
				utils.ExitWithErrorMsg("ERR: specified path is not a directory:", nodeHomeDirectory)
				return
			}

			dataDirPath := path.Join(nodeHomeDirectory, "data")
			_, exists, isDir, err = utils.FileInfo(dataDirPath)
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
			if exists {
				utils.ExitWithErrorMsg("ERR:", dataDirPath, "is not empty, require reset data")
				return
			}

			configDirPath := path.Join(nodeHomeDirectory, "config")
			_, exists, isDir, err = utils.FileInfo(configDirPath)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check config directory:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: config directory does not exist:", configDirPath)
				return
			}
			if !isDir {
				utils.ExitWithErrorMsg("ERR: required config path is not a directory:", configDirPath)
				return
			}

			if binary == "" {
				utils.ExitWithErrorMsgf("ERR: required flag --%s\n", flagBinary)
				return
			}
			if err := validation.ValidateNodeBinary(binary); err != nil {
				utils.ExitWithErrorMsg("ERR:", err.Error())
				return
			}
			if !allowLocalBinary {
				if !strings.Contains(binary, "/") {
					utils.ExitWithErrorMsg("ERR:", msgDescFlagAllowLocalBinary)
					return
				}
				currentUser, err := user.Current()
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to get current user")
					return
				}
				currentUserHome := strings.TrimSuffix(currentUser.HomeDir, "/")
				binaryHome, err := utils.TryExtractUserHomeDirFromPath(binary)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to extract binary user home dir from path", binary)
					return
				}
				binaryHome = strings.TrimSuffix(binaryHome, "/")
				if binaryHome == currentUserHome {
					utils.ExitWithErrorMsgf("ERR: supplied binary located in the same home dir as current user, use flag --%s to allow\n", flagAllowLocalBinary)
					return
				}
			}

			if newPeers != "" && !validation.IsValidPeer(newPeers) {
				utils.ExitWithErrorMsg("ERR: provided peers is invalid:", newPeers)
				return
			}
			if seeds != "" && !validation.IsValidPeer(seeds) {
				utils.ExitWithErrorMsg("ERR: provided seeds is invalid:", seeds)
				return
			}

			if addressBookFilePath != "" {
				_, exists, isDir, err = utils.FileInfo(addressBookFilePath)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to check provided address book file:", err)
					return
				}
				if !exists {
					utils.ExitWithErrorMsg("ERR: provided address book file does not exist:", addressBookFilePath)
					return
				}
				if isDir {
					utils.ExitWithErrorMsg("ERR: provided address book path is a directory:", addressBookFilePath)
					return
				}

				addrBook, err := readAddrBook(addressBookFilePath)
				if err != nil {
					utils.ExitWithErrorMsg("ERR: failed to read provided address book file:", err)
					return
				}

				livePeers := getLivePeers(addrBook, 5*time.Hour)

				if len(livePeers) == 0 {
					utils.PrintlnStdErr("WARN: no live peers found in provided address book")
				} else {
					fmt.Println("INF: found", len(livePeers), "live peers in provided address book")

					if newPeers != "" {
						newPeers += ","
					}
					for i, livePeer := range livePeers {
						if i > 0 {
							newPeers += ","
						}
						newPeers += fmt.Sprintf("%s@%s:%d", livePeer.Addr.ID, livePeer.Addr.IP.String(), livePeer.Addr.Port)
					}
				}
			}

			configFilePath := path.Join(configDirPath, "config.toml")
			_, exists, _, err = utils.FileInfo(configFilePath)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to check config file:", err)
				return
			}
			if !exists {
				utils.ExitWithErrorMsg("ERR: config file does not exist:", configFilePath)
				return
			}

			var modernSed bool
			launchSed := func(pattern string) {
				args := []string{"-i"}
				if utils.IsDarwin() {
					args = append(args, "")
				}
				if modernSed {
					args = append(args, "-E")
				} else {
					args = append(args, "-e")
				}
				args = append(args, pattern, configFilePath)

				ec := utils.LaunchApp("sed", args)
				if ec != 0 {
					utils.PrintlnStdErr("ERR: failed to launch sed to update config file")
					utils.PrintlnStdErr("> sed " + strings.Join(args, " "))
					os.Exit(ec)
				}
			}

			modernSed = false
			if seeds != "" {
				launchSed(fmt.Sprintf("s/^seeds *=.*/seeds = \"%s\"/", seeds))
				fmt.Println("INF: seeds updated in config file")
			}
			if newPeers != "" {
				launchSed(fmt.Sprintf("s/^persistent_peers *=.*/persistent_peers = \"%s\"/", newPeers))
				fmt.Println("INF: persistent_peers updated in config file")
			}

			rpc = strings.TrimSuffix(rpc, "/")

			output, ec := utils.LaunchAppAndGetOutput("/bin/bash", []string{"-c", fmt.Sprintf("curl -s %s/block | jq -r .result.block.header.height", rpc)})
			if ec != 0 {
				utils.PrintlnStdErr(output)
				utils.ExitWithErrorMsg("ERR: failed to get block height from rpc:", rpc)
				return
			}
			blockHeight, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: failed to parse block height", output, "from rpc:", err)
				return
			}
			if blockHeight > 7000 {
				blockHeight = blockHeight - 2000
			} else if blockHeight >= 500 {
				blockHeight = blockHeight / 100 * 100
			}

			fmt.Println("Block height:", blockHeight)

			output, ec = utils.LaunchAppAndGetOutput("/bin/bash", []string{"-c", fmt.Sprintf(`curl -s "%s/block?height=%d" | jq -r .result.block_id.hash`, rpc, blockHeight)})
			if ec != 0 {
				utils.PrintlnStdErr(output)
				utils.ExitWithErrorMsg("ERR: failed to get block hash from rpc:", rpc)
				return
			}
			trustHash := strings.TrimSpace(output)
			if !regexp.MustCompile(`^[A-F\d]{64}$`).MatchString(trustHash) {
				utils.ExitWithErrorMsg("ERR: invalid block hash", trustHash, "from rpc")
				return
			}

			modernSed = true
			launchSed(`s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"` + rpc + "," + rpc + `\"| ; s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1` + fmt.Sprintf("%d", blockHeight) + `| ; s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1"` + trustHash + `"|`)
			fmt.Println("INF: trust_height, rpc_servers, trust_hash and enable are updated in config file")

			launchCmd := exec.Command(binary, []string{"start", "--home", nodeHomeDirectory}...)
			launchCmd.Stdout = os.Stdout
			launchCmd.Stderr = os.Stderr
			err = launchCmd.Start()
			if err != nil {
				utils.ExitWithErrorMsgf("ERR: %s state-sync failed to start %v\n", binary, err)
				return
			}

			expiry := time.Now().UTC().Add(maxDuration)
			for {
				if time.Now().UTC().After(expiry) {
					utils.ExitWithErrorMsg("ERR: state sync expired")
					return
				}

				time.Sleep(30 * time.Second)

				output, ec = utils.LaunchAppAndGetOutput("/bin/bash", []string{"-c", fmt.Sprintf("curl -s %s/status | jq -r .result.sync_info.catching_up", rpc)})
				if ec != 0 {
					utils.PrintlnStdErr("ERR: failed to get catching_up from rpc")
					continue
				}

				catchingUp := strings.ToLower(strings.TrimSpace(output))
				switch catchingUp {
				case "false":
					fmt.Println("INF: node is synced")
					_ = launchCmd.Process.Kill()
					os.Exit(0)
				case "true":
					fmt.Println("INF: node is catching up")
					continue
				default:
					utils.PrintlnStdErr("ERR: invalid catching_up value from rpc:", catchingUp)
					continue
				}
			}
		},
	}

	cmd.Flags().String(flagBinary, "", "Path to the binary")
	cmd.Flags().Bool(flagAllowLocalBinary, false, "By default, "+msgDescFlagAllowLocalBinary)
	cmd.Flags().String(flagAddressBook, "", "Path to the address book file to take live peers from")
	cmd.Flags().String(flagPeers, "", "List of peers to use for state sync")
	cmd.Flags().String(flagSeeds, "", "List of seeds to use for state sync")
	cmd.Flags().String(flagRpc, "http://localhost:26657", "RPC address to use for state sync")
	cmd.Flags().Duration(flagMaxDuration, 12*time.Hour, "Maximum duration to wait for state sync")

	return cmd
}
