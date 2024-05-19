package web_server

import (
	"github.com/bcdevtools/node-management/types"
	"github.com/gin-gonic/gin"
	"time"
)

var cacheSnapshotData *types.TimeBasedCache

func HandleWebPage(c *gin.Context) {
	//w := wrapGin(c)

	//w.PrepareDefaultSuccessResponse(peers).SendResponse()
}

func init() {
	cacheSnapshotData = types.NewTimeBasedCache(60 * time.Second)
}
