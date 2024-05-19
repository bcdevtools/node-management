package web_server

import (
	"fmt"
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/services/web_server/gin_wrapper"
	webtypes "github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	statikfs "github.com/rakyll/statik/fs"
	"html/template"
	"net/http"
)

func StartWebServer(cfg webtypes.Config) {
	if err := validation.PossibleNodeHome(cfg.NodeHome); err != nil {
		utils.PrintlnStdErr("ERR: invalid node home directory:", err)
		return
	}

	binding := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	statikFS, err := statikfs.New()
	if err != nil {
		panic(errors.Wrap(err, "failed to create statik FS"))
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(constants.GinConfig, cfg)
	})

	const (
		engineDelimsLeft  = "{[{"
		engineDelimsRight = "}]}"
	)
	r.Delims(engineDelimsLeft, engineDelimsRight)
	r.SetHTMLTemplate(
		template.Must(
			template.
				New("").
				Delims(engineDelimsLeft, engineDelimsRight).
				Funcs(nil).
				ParseFS(
					webtypes.WrapHttpFsToOsFs(statikFS),
					"/index.tmpl",
				),
		),
	)

	// Resources
	r.GET("/resources/*file", func(c *gin.Context) {
		http.FileServer(statikFS).ServeHTTP(c.Writer, c.Request)
	})

	// API
	r.GET("/api/node/live-peers", HandleApiNodeLivePeers)

	// Web
	r.GET("/", HandleWebIndex)

	fmt.Println("INF: starting Web service at", binding)

	if err := r.Run(binding); err != nil {
		utils.PrintlnStdErr("ERR: failed to start Web service")
		panic(err)
	}
}

// wrap and return gin Context as a GinWrapper class with enhanced utilities
func wrapGin(c *gin.Context) gin_wrapper.GinWrapper {
	return gin_wrapper.WrapGin(c)
}
