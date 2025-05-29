package web_server

import (
	webtypes "github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

var cacheAddrBook *types.TimeBasedCache

func HandleDownloadAddrBook(c *gin.Context) {
	w := wrapGin(c)

	addrBook, err := getAddrbook(w.Config())
	if err != nil {
		utils.PrintlnStdErr("ERR: failed to get addrbook.json:", err)
		w.PrepareDefaultErrorResponse().
			WithResult("failed to get addrbook.json").
			SendResponse()
		return
	}
	if addrBook == nil || len(addrBook.Addrs) == 0 {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusServiceUnavailable).
			WithResult("failed to get addrbook.json").
			SendResponse()
		return
	}

	c.Header("Content-Disposition", "attachment; filename=addrbook.json")
	c.JSON(http.StatusOK, addrBook)
}

func getAddrbook(cfg webtypes.Config) (*types.AddrBook, error) {
	if addrBook := cacheAddrBook.GetRL(); addrBook != nil {
		return addrBook.(*types.AddrBook), nil
	}

	addrBook, err := cacheAddrBook.UpdateWL(func() (any, error) {
		addrBook := &types.AddrBook{}
		if err := addrBook.ReadAddrBook(cfg.GetAddrBookFilePath()); err != nil {
			return nil, errors.Wrap(err, "failed to read addrbook")
		}

		livePeers := addrBook.GetLivePeers(48*time.Hour, false)

		if len(livePeers) == 0 && cfg.Debug {
			// load random, include dead peers, on debug mode
			livePeers = addrBook.Addrs
			if len(livePeers) > 10 {
				livePeers = livePeers[:10]
			}
		}

		addrBook.Addrs = livePeers

		return addrBook, nil
	}, true)

	if err != nil {
		return nil, err
	}

	return addrBook.(*types.AddrBook), nil
}

func init() {
	cacheAddrBook = types.NewTimeBasedCache(60 * time.Second)
}
