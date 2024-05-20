package cmd

import (
	"fmt"
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

func GetGenStartWebCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "gen-start-web",
		Short: fmt.Sprintf("Generate the `%s %s` command", constants.BINARY_NAME, cmdStartWeb),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Referral node home, for getting live-peers, (eg: ~/.gaia):")

			nodeHomeDirectory := strings.TrimSuffix(strings.TrimSpace(utils.ReadText(false)), "/")
			err := validation.PossibleNodeHome(nodeHomeDirectory)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: invalid node home directory:", err)
				return
			}

			fmt.Printf("\nPort (default %d):\n", defaultWebPort)
			port := utils.ReadOptionalNumber(1023, 65535, defaultWebPort)

			fmt.Println("Brand (eg: valoper.io):")
			brand := strings.TrimSpace(utils.ReadText(false))
			if brand == "" {
				utils.ExitWithErrorMsg("ERR: brand is required")
				return
			}
			if strings.Contains(brand, "\"") || strings.Contains(brand, "'") {
				utils.ExitWithErrorMsg("ERR: brand must not contain double or single quotes")
				return
			}

			fmt.Println("Authorization token:")
			authorizationToken := strings.TrimSpace(utils.ReadText(false))
			if authorizationToken == "" {
				utils.ExitWithErrorMsg("ERR: authorization token is required")
				return
			}
			if !regexp.MustCompile(`^[a-zA-Z\d_-]{16,}$`).MatchString(authorizationToken) {
				utils.ExitWithErrorMsg("ERR: authorization token must be at least 16 characters long and contain only letters, digits, hyphens/dash, and underscores")
				return
			}

			fmt.Println("Chain name (eg: Cosmos Hub):")
			chainName := strings.TrimSpace(utils.ReadText(false))
			if chainName == "" {
				utils.ExitWithErrorMsg("ERR: chain name is required")
				return
			}
			if !regexp.MustCompile(`^[a-zA-Z\d\s_-]+$`).MatchString(chainName) {
				utils.ExitWithErrorMsg("ERR: chain name must contain only letters, digits, spaces, hyphens/dash, and underscores")
				return
			}

			fmt.Println("Chain ID (eg: cosmoshub-4):")
			chainID := strings.TrimSpace(utils.ReadText(false))
			if chainID == "" {
				utils.ExitWithErrorMsg("ERR: chain ID is required")
				return
			}
			if !regexp.MustCompile(`^[a-zA-Z\d_-]+$`).MatchString(chainID) {
				utils.ExitWithErrorMsg("ERR: chain ID must contain only letters, digits, hyphens/dash, and underscores")
				return
			}

			fmt.Println("General binary name (eg: gaiad):")
			generalBinaryName := strings.TrimSpace(utils.ReadText(false))
			if generalBinaryName == "" {
				utils.ExitWithErrorMsg("ERR: general binary name is required")
				return
			}
			if !regexp.MustCompile(`^[a-z\d_-]+$`).MatchString(generalBinaryName) {
				utils.ExitWithErrorMsg("ERR: general binary name must contain only lowercase letters, digits, hyphens/dash, and underscores")
				return
			}

			fmt.Println("General node home name (eg: .gaia):")
			generalNodeHomeName := strings.TrimSuffix(strings.TrimSpace(utils.ReadText(false)), "/")
			if generalNodeHomeName == "" {
				utils.ExitWithErrorMsg("ERR: general node home name is required")
				return
			}
			if !strings.HasPrefix(generalNodeHomeName, ".") {
				utils.ExitWithErrorMsg("ERR: general node home name must start with a dot")
				return
			}
			if strings.Contains(generalNodeHomeName, "/") {
				utils.ExitWithErrorMsg("ERR: general node home name must be name only, not a path")
				return
			}
			if !regexp.MustCompile(`^\.[a-zA-Z\d_-]+$`).MatchString(generalNodeHomeName) {
				utils.ExitWithErrorMsg("ERR: general node home name must contain only letters, digits, hyphens/dash, and underscores")
				return
			}

			fmt.Println("Snapshot file path on machine (eg: /snapshot/snapshot_cosmoshub.tar.lz4):")
			snapshotFilePath := strings.TrimSuffix(strings.TrimSpace(utils.ReadText(false)), "/")
			if snapshotFilePath == "" {
				utils.ExitWithErrorMsg("ERR: snapshot file path is required")
				return
			}
			if !strings.HasPrefix(snapshotFilePath, "/") {
				utils.ExitWithErrorMsg("ERR: snapshot file path must be absolute")
				return
			}
			if !strings.HasSuffix(snapshotFilePath, ".tar.lz4") {
				utils.ExitWithErrorMsg("ERR: snapshot file path must be a .tar.lz4 file")
				return
			}
			for {
				perm, exists, _, err := utils.FileInfo(snapshotFilePath)
				if err != nil {
					err = errors.Wrap(err, "failed to check snapshot file")
				} else if !exists {
					utils.ExitWithErrorMsg("ERR: snapshot file does not exist")
					return
				} else {
					filePerm := types.FilePermFrom(perm)
					if !filePerm.Other.Read || !filePerm.Group.Read || !filePerm.User.Read {
						err = fmt.Errorf("lacking read permission for snapshot file")
					} else if filePerm.Other.Write || filePerm.Group.Write {
						err = fmt.Errorf("unnecessary write permission for snapshot file")
					}
				}

				if err == nil {
					break
				}

				utils.PrintlnStdErr("ERR:", err.Error())
				utils.PrintlnStdErr("Please correct the snapshot file path or permission!")
				utils.PrintlnStdErr("Command to fix snapshot file permission:")
				utils.PrintlnStdErr("> sudo chmod ugo+r", snapshotFilePath)
				utils.PrintlnStdErr("> sudo chmod go-w", snapshotFilePath)
				utils.PrintlnStdErr("Then press enter to continue...")
				_ = utils.ReadText(true)
			}

			fmt.Println("Snapshot download URL (eg: https://example.com/snapshot_cosmoshub.tar.lz4):")
			snapshotDownloadUrl := strings.TrimSpace(utils.ReadText(false))
			if snapshotDownloadUrl == "" {
				utils.ExitWithErrorMsg("ERR: snapshot download URL is required")
				return
			}
			if //goland:noinspection HttpUrlsUsage
			!strings.HasPrefix(snapshotDownloadUrl, "http://") && !strings.HasPrefix(snapshotDownloadUrl, "https://") {
				utils.ExitWithErrorMsg("ERR: snapshot download URL must contain protocol")
				return
			}

			fmt.Println("RPC URL (eg: https://rpc1.cosmos.m.example.com):")
			externalRpcUrl := strings.TrimSuffix(strings.TrimSpace(utils.ReadText(false)), "/")
			if externalRpcUrl == "" {
				utils.ExitWithErrorMsg("ERR: RPC URL is required")
				return
			}
			if !strings.Contains(externalRpcUrl, "://") {
				utils.ExitWithErrorMsg("ERR: RPC URL must contain protocol")
				return
			}
			if strings.Contains(externalRpcUrl, "localhost") || strings.Contains(externalRpcUrl, "127.0.0.1") {
				utils.ExitWithErrorMsg("ERR: RPC URL must not be localhost")
				return
			}

			fmt.Println("REST URL (eg: https://rest1.cosmos.m.example.com):")
			externalRestUrl := strings.TrimSuffix(strings.TrimSpace(utils.ReadText(false)), "/")
			if externalRestUrl == "" {
				utils.ExitWithErrorMsg("ERR: REST URL is required")
				return
			}
			if !strings.Contains(externalRestUrl, "://") {
				utils.ExitWithErrorMsg("ERR: REST URL must contain protocol")
				return
			}
			if strings.Contains(externalRestUrl, "localhost") || strings.Contains(externalRestUrl, "127.0.0.1") {
				utils.ExitWithErrorMsg("ERR: REST URL must not be localhost")
				return
			}
			if externalRestUrl == externalRpcUrl {
				utils.ExitWithErrorMsg("ERR: REST URL must be different from RPC URL")
				return
			}

			fmt.Println("gRPC URL (optional, eg: https://grpc1.cosmos.m.example.com):")
			externalGrpcUrl := strings.TrimSuffix(strings.TrimSpace(utils.ReadText(true)), "/")
			if externalGrpcUrl != "" {
				if !strings.Contains(externalGrpcUrl, "://") {
					utils.ExitWithErrorMsg("ERR: gRPC URL must contain protocol")
					return
				}
				if strings.Contains(externalGrpcUrl, "localhost") || strings.Contains(externalGrpcUrl, "127.0.0.1") {
					utils.ExitWithErrorMsg("ERR: gRPC URL must not be localhost")
					return
				}
			}

			fmt.Println("External resource logo URL (optional, eg: https://example.com/logo.png):")
			extResLogoUrl := strings.TrimSpace(utils.ReadText(true))

			fmt.Println("External resource favicon URL (optional, eg: https://example.com/favicon.ico):")
			extResFaviconUrl := strings.TrimSpace(utils.ReadText(true))

			var sb strings.Builder
			{
				sb.WriteString(constants.BINARY_NAME)
				sb.WriteString(" ")
				sb.WriteString(cmdStartWeb)
				sb.WriteString(" ")
				sb.WriteString(nodeHomeDirectory)
			}
			if port != defaultWebPort {
				sb.WriteString(" --")
				sb.WriteString(flagPort)
				sb.WriteString(" ")
				sb.WriteString(fmt.Sprintf("%d", port))
			}
			if !strings.EqualFold(brand, defaultBrand) {
				sb.WriteString(" --")
				sb.WriteString(flagBrand)
				sb.WriteString(" '")
				sb.WriteString(brand)
				sb.WriteString("'")
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagAuthorizationToken)
				sb.WriteString(" '")
				sb.WriteString(authorizationToken)
				sb.WriteString("'")
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagChainName)
				sb.WriteString(" '")
				sb.WriteString(chainName)
				sb.WriteString("'")
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagChainID)
				sb.WriteString(" ")
				sb.WriteString(chainID)
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagGeneralBinaryName)
				sb.WriteString(" ")
				sb.WriteString(generalBinaryName)
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagGeneralNodeHomeName)
				sb.WriteString(" ")
				sb.WriteString(generalNodeHomeName)
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagSnapshotFilePath)
				sb.WriteString(" ")
				sb.WriteString(snapshotFilePath)
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagSnapshotDownloadURL)
				sb.WriteString(" ")
				sb.WriteString(snapshotDownloadUrl)
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagExtResRpcUrl)
				sb.WriteString(" ")
				sb.WriteString(externalRpcUrl)
			}
			{
				sb.WriteString(" --")
				sb.WriteString(flagExtResRestUrl)
				sb.WriteString(" ")
				sb.WriteString(externalRestUrl)
			}
			if externalGrpcUrl != "" {
				sb.WriteString(" --")
				sb.WriteString(flagExtResGrpcUrl)
				sb.WriteString(" ")
				sb.WriteString(externalGrpcUrl)
			}
			if extResLogoUrl != "" {
				sb.WriteString(" --")
				sb.WriteString(flagExtResLogoUrl)
				sb.WriteString(" ")
				sb.WriteString(extResLogoUrl)
			}
			if extResFaviconUrl != "" {
				sb.WriteString(" --")
				sb.WriteString(flagExtResFaviconUrl)
				sb.WriteString(" ")
				sb.WriteString(extResFaviconUrl)
			}

			fmt.Println()
			fmt.Println("Generated command:")
			fmt.Println()
			fmt.Println(sb.String())
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(GetGenStartWebCmd())
}
