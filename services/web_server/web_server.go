package web_server

import (
	"fmt"
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/services/web_server/gin_wrapper"
	"github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/gin-gonic/gin"
)

func StartWebServer(cfg types.Config) {
	if err := validation.PossibleNodeHome(cfg.NodeHome); err != nil {
		utils.PrintlnStdErr("ERR: invalid node home directory:", err)
		return
	}

	binding := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(constants.GinConfig, cfg)
	})

	// API
	r.GET("/api/node/live-peers", HandleApiNodeLivePeers)

	fmt.Println("INF: starting Web service at", binding)

	err := r.Run(binding)
	if err != nil {
		utils.PrintlnStdErr("ERR: failed to start Web service")
		panic(err)
	}
}

// wrap and return gin Context as a GinWrapper class with enhanced utilities
func wrapGin(c *gin.Context) gin_wrapper.GinWrapper {
	return gin_wrapper.WrapGin(c)
}
