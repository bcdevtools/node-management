package gin_wrapper

import (
	"github.com/bcdevtools/node-management/constants"
	"github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/gin-gonic/gin"
)

type ginWrapperType int8

const (
	GwtDefault ginWrapperType = iota
)

type GinWrapper struct {
	c     *gin.Context
	wType ginWrapperType
}

func WrapGin(c *gin.Context) GinWrapper {
	return GinWrapper{
		c:     c,
		wType: GwtDefault,
	}
}

func (w GinWrapper) Gin() *gin.Context {
	return w.c
}

func (w GinWrapper) Binder() *GinBinder {
	return &GinBinder{
		c:   w.c,
		err: nil,
	}
}

func (w GinWrapper) Config() types.Config {
	return w.c.MustGet(constants.GinConfig).(types.Config)
}

func (w GinWrapper) IsAuthorizedRequest() bool {
	cfg := w.Config()
	if cfg.AuthorizeToken == "" {
		return false
	}

	token := w.c.GetHeader("VN-Authorization")
	if token == "" {
		return false
	}

	return token == cfg.AuthorizeToken
}
