package cmd

import (
	"github.com/bcdevtools/node-management/services/web_server"
	apitypes "github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/spf13/cobra"
	"strings"
)

const (
	flagPort               = "port"
	flagAuthorizationToken = "authorization-token"
	flagDebug              = "debug"
)

func GetStartWebCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start-web [node_home]",
		Short: "Start Web service",
		Run: func(cmd *cobra.Command, args []string) {
			utils.MustNotUserRoot()

			nodeHomeDirectory := strings.TrimSpace(args[0])
			port, _ := cmd.Flags().GetUint16(flagPort)
			authorizationToken, _ := cmd.Flags().GetString(flagAuthorizationToken)
			debug, _ := cmd.Flags().GetBool(flagDebug)

			err := validation.PossibleNodeHome(nodeHomeDirectory)
			if err != nil {
				utils.ExitWithErrorMsg("ERR: invalid node home directory:", err)
				return
			}

			authorizationToken = strings.TrimSpace(authorizationToken)
			if authorizationToken == "" {
				utils.ExitWithErrorMsg("ERR: authorization token is required")
				return
			}

			web_server.StartWebServer(apitypes.Config{
				Port:           port,
				AuthorizeToken: authorizationToken,
				NodeHome:       nodeHomeDirectory,
				Debug:          debug,
			})
		},
	}

	cmd.Flags().Uint16(flagPort, 8080, "port to bind Web service to")
	cmd.Flags().StringP(flagAuthorizationToken, "a", "", "authorization token")
	cmd.Flags().Bool(flagDebug, false, "enable debug mode")

	return cmd
}

func init() {
	rootCmd.AddCommand(GetStartWebCmd())
}
