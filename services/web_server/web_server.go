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
	"strings"
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
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(strings.TrimPrefix(c.Request.URL.Path, "/"), "api/internal") {
			w := wrapGin(c)
			if !w.IsAuthorizedRequest() {
				w.PrepareDefaultErrorResponse().
					WithHttpStatusCode(http.StatusForbidden).
					WithResult("invalid authentication token").
					SendResponse()
				c.Abort()
				return
			}
		}

		c.Next()
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
	r.GET("/api/internal/monitoring/stats", HandleApiInternalMonitoringStats)

	// Web
	r.GET("/", HandleWebIndex)
	r.GET("/download/addrbook.json", HandleDownloadAddrBook)

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
