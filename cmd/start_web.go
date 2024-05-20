package cmd

import (
	_ "github.com/bcdevtools/node-management/client/statik"
	"github.com/bcdevtools/node-management/services/web_server"
	webtypes "github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/spf13/cobra"
	"strings"
)

const (
	flagPort               = "port"
	flagAuthorizationToken = "authorization-token"
	flagDebug              = "debug"

	flagBrand               = "brand"
	flagChainName           = "chain-name"
	flagChainID             = "chain-id"
	flagGeneralBinaryName   = "g-binary-name"
	flagGeneralNodeHomeName = "g-node-home-name"

	flagExtResLogoUrl    = "exr-logo-url"
	flagExtResFaviconUrl = "exr-favicon-url"
	flagExtResRpcUrl     = "exr-rpc-url"
	flagExtResRestUrl    = "exr-rest-url"
	flagExtResGrpcUrl    = "exr-grpc-url"

	flagSnapshotFilePath    = "snapshot-file"
	flagSnapshotDownloadURL = "snapshot-download-url"
)

const (
	cmdStartWeb = "start-web"
)

const (
	defaultWebPort = 8080
	defaultBrand   = "Valoper.io"
)

func GetStartWebCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   cmdStartWeb + " [node_home]",
		Short: "Start Web service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			nodeHomeDirectory := strings.TrimSpace(args[0])
			port, _ := cmd.Flags().GetUint16(flagPort)
			authorizationToken, _ := cmd.Flags().GetString(flagAuthorizationToken)
			debug, _ := cmd.Flags().GetBool(flagDebug)

			brand, _ := cmd.Flags().GetString(flagBrand)
			chainName, _ := cmd.Flags().GetString(flagChainName)
			chainID, _ := cmd.Flags().GetString(flagChainID)
			generalBinaryName, _ := cmd.Flags().GetString(flagGeneralBinaryName)
			generalNodeHomeName, _ := cmd.Flags().GetString(flagGeneralNodeHomeName)

			extResLogoUrl, _ := cmd.Flags().GetString(flagExtResLogoUrl)
			extResFaviconUrl, _ := cmd.Flags().GetString(flagExtResFaviconUrl)
			extResRpcUrl, _ := cmd.Flags().GetString(flagExtResRpcUrl)
			extResRestUrl, _ := cmd.Flags().GetString(flagExtResRestUrl)
			extResGrpcUrl, _ := cmd.Flags().GetString(flagExtResGrpcUrl)

			snapshotFilePath, _ := cmd.Flags().GetString(flagSnapshotFilePath)
			snapshotDownloadURL, _ := cmd.Flags().GetString(flagSnapshotDownloadURL)

			err := validation.PossibleNodeHome(nodeHomeDirectory)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: invalid node home directory:", err)
				return
			}

			authorizationToken = strings.TrimSpace(authorizationToken)
			if authorizationToken == "" {
				utils.ExitWithErrorMsgf("ERR: authorization token is required, use --%s flag to set it\n", flagAuthorizationToken)
				return
			}

			brand = strings.TrimSpace(brand)
			if brand == "" {
				utils.ExitWithErrorMsgf("ERR: brand is required, use --%s flag to set it\n", flagBrand)
				return
			}

			chainName = strings.TrimSpace(chainName)
			if chainName == "" {
				utils.ExitWithErrorMsgf("ERR: chain name is required, use --%s flag to set it\n", flagChainName)
				return
			}

			chainID = strings.TrimSpace(chainID)
			if chainID == "" {
				utils.ExitWithErrorMsgf("ERR: chain ID is required, use --%s flag to set it\n", flagChainID)
				return
			}

			generalBinaryName = strings.TrimSpace(generalBinaryName)
			if generalBinaryName == "" {
				utils.ExitWithErrorMsgf("ERR: general binary name is required, use --%s flag to set it\n", flagGeneralBinaryName)
				return
			}

			generalNodeHomeName = strings.TrimSpace(generalNodeHomeName)
			if generalNodeHomeName == "" {
				utils.ExitWithErrorMsgf("ERR: general node home name is required, use --%s flag to set it\n", flagGeneralNodeHomeName)
				return
			}
			if !strings.HasPrefix(generalNodeHomeName, ".") {
				utils.ExitWithErrorMsgf("ERR: general node home name must starts with a dot, correct the --%s flag\n", flagGeneralNodeHomeName)
				return
			}
			if strings.Contains(generalNodeHomeName, "/") {
				utils.ExitWithErrorMsgf("ERR: general node home name must be name only, not a path, correct the --%s flag\n", flagGeneralNodeHomeName)
				return
			}

			extResLogoUrl = strings.TrimSpace(extResLogoUrl)
			extResFaviconUrl = strings.TrimSpace(extResFaviconUrl)
			extResRpcUrl = strings.TrimSpace(extResRpcUrl)
			extResRestUrl = strings.TrimSpace(extResRestUrl)
			extResGrpcUrl = strings.TrimSpace(extResGrpcUrl)

			if extResLogoUrl == "" && extResFaviconUrl != "" {
				extResLogoUrl = extResFaviconUrl
			}

			if extResRpcUrl == "" {
				utils.ExitWithErrorMsgf("ERR: external resource RPC URL is required, use --%s flag to set it\n", flagExtResRpcUrl)
				return
			}
			if !strings.Contains(extResRpcUrl, "://") || strings.Contains(extResRpcUrl, "localhost") || strings.Contains(extResRpcUrl, "127.0.0.1") {
				utils.ExitWithErrorMsgf("ERR: external resource RPC URL must contains protocol, not localhost, correct the --%s flag\n", flagExtResRpcUrl)
				return
			}
			if extResRestUrl == "" {
				utils.ExitWithErrorMsgf("ERR: external resource REST URL is required, use --%s flag to set it\n", flagExtResRestUrl)
				return
			}
			if !strings.Contains(extResRestUrl, "://") || strings.Contains(extResRestUrl, "localhost") || strings.Contains(extResRestUrl, "127.0.0.1") {
				utils.ExitWithErrorMsgf("ERR: external resource REST URL must contains protocol, not localhost, correct the --%s flag\n", flagExtResRestUrl)
				return
			}
			if extResGrpcUrl != "" {
				if !strings.Contains(extResGrpcUrl, "://") || strings.Contains(extResGrpcUrl, "localhost") || strings.Contains(extResGrpcUrl, "127.0.0.1") {
					utils.ExitWithErrorMsgf("ERR: external resource gRPC URL must contains protocol, not localhost, correct the --%s flag\n", flagExtResGrpcUrl)
					return
				}
			}

			snapshotFilePath = strings.TrimSpace(snapshotFilePath)
			if snapshotFilePath == "" {
				utils.ExitWithErrorMsgf("ERR: snapshot file path is required, use --%s flag to set it\n", flagSnapshotFilePath)
				return
			}

			snapshotDownloadURL = strings.TrimSuffix(strings.TrimSpace(snapshotDownloadURL), "/")
			if snapshotDownloadURL == "" {
				utils.ExitWithErrorMsgf("ERR: snapshot download URL is required, use --%s flag to set it\n", flagSnapshotDownloadURL)
				return
			}

			web_server.StartWebServer(webtypes.Config{
				Port:           port,
				AuthorizeToken: authorizationToken,
				NodeHome:       nodeHomeDirectory,
				Debug:          debug,

				Brand: brand,

				ChainName:           chainName,
				ChainID:             chainID,
				GeneralBinaryName:   generalBinaryName,
				GeneralNodeHomeName: generalNodeHomeName,

				ExternalResourceLogoUrl:    extResLogoUrl,
				ExternalResourceFaviconUrl: extResFaviconUrl,
				ExternalResourceRpcUrl:     extResRpcUrl,
				ExternalResourceRestUrl:    extResRestUrl,
				ExternalResourceGrpcUrl:    extResGrpcUrl,

				SnapshotFilePath:    snapshotFilePath,
				SnapshotDownloadURL: snapshotDownloadURL,
			})
		},
	}

	cmd.Flags().Uint16(flagPort, defaultWebPort, "port to bind Web service to")
	cmd.Flags().StringP(flagAuthorizationToken, "a", "", "authorization token")
	cmd.Flags().Bool(flagDebug, false, "enable debug mode")

	cmd.Flags().String(flagBrand, defaultBrand, "brand")
	cmd.Flags().String(flagChainName, "", "chain name")
	cmd.Flags().String(flagChainID, "", "chain ID")
	cmd.Flags().String(flagGeneralBinaryName, "", "general binary name")
	cmd.Flags().String(flagGeneralNodeHomeName, "", "general node home name")

	cmd.Flags().String(flagExtResLogoUrl, "", "external resource logo URL")
	cmd.Flags().String(flagExtResFaviconUrl, "", "external resource favicon URL")
	cmd.Flags().String(flagExtResRpcUrl, "", "external node RPC URL, used to write guide")
	cmd.Flags().String(flagExtResRestUrl, "", "external node REST URL, used to write guide")
	cmd.Flags().String(flagExtResGrpcUrl, "", "external node gRPC URL, used to write guide")

	cmd.Flags().String(flagSnapshotFilePath, "", "snapshot local file path")
	cmd.Flags().String(flagSnapshotDownloadURL, "", "snapshot download URL")

	return cmd
}

func init() {
	rootCmd.AddCommand(GetStartWebCmd())
}
