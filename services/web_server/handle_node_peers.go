package web_server

import (
	"fmt"
	webtypes "github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"slices"
	"time"
)

var cacheNodePeers *types.TimeBasedCache

func HandleApiNodeLivePeers(c *gin.Context) {
	w := wrapGin(c)

	peers, err := getLivePeers(w.Config())
	if err != nil {
		utils.PrintlnStdErr("ERR: failed to get live peers:", err)
		w.PrepareDefaultErrorResponse().WithResult("failed to get live peers").SendResponse()
		return
	}

	w.PrepareDefaultSuccessResponse(peers).SendResponse()
}

func getLivePeers(cfg webtypes.Config) ([]string, error) {
	if peers := cacheNodePeers.GetRL(); peers != nil {
		return peers.([]string), nil
	}

	peers, err := cacheNodePeers.UpdateWL(func() (any, error) {
		addrBook := &types.AddrBook{}
		if err := addrBook.ReadAddrBook(cfg.GetAddrBookFilePath()); err != nil {
			return nil, errors.Wrap(err, "failed to read addrbook")
		}

		livePeers := addrBook.GetLivePeers(1 * time.Hour)

		if len(livePeers) == 0 && cfg.Debug {
			// load random, include dead peers, on debug mode
			livePeers = addrBook.Addrs
			if len(livePeers) > 10 {
				livePeers = livePeers[:10]
			}
		}

		slices.SortFunc(livePeers, func(left, right *types.KnownAddress) int {
			leftConnected := !left.LastAttempt.After(left.LastSuccess)
			rightConnected := !right.LastAttempt.After(right.LastSuccess)
			if leftConnected == rightConnected {
				return 0
			}
			if leftConnected {
				return -1
			} else {
				return 1
			}
		})

		peers := make([]string, len(livePeers))

		for i, peer := range livePeers {
			peers[i] = fmt.Sprintf("%s@%s:%d", peer.Addr.ID, peer.Addr.IP, peer.Addr.Port)
		}

		return peers, nil
	}, true)

	if err != nil {
		return nil, err
	}

	return peers.([]string), nil
}

func init() {
	cacheNodePeers = types.NewTimeBasedCache(60 * time.Second)
}
