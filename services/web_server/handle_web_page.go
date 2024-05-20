package web_server

import (
	"fmt"
	webtypes "github.com/bcdevtools/node-management/services/web_server/types"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var cacheSnapshotData *types.TimeBasedCache

func HandleWebIndex(c *gin.Context) {
	w := wrapGin(c)
	cfg := w.Config()

	var livePeers string
	var livePeersCount int

	peers, err := getLivePeers(w.Config())
	if err != nil {
		utils.PrintlnStdErr("ERR: failed to get live peers:", err)
	} else {
		const maximumPeers = 90
		if len(peers) > maximumPeers {
			peers = peers[:maximumPeers]
		}
		livePeers = strings.Join(peers, ",")
		livePeersCount = len(peers)
	}

	snapshotInfo := getSnapshotInfo(cfg)

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":               fmt.Sprintf("%s snapshot by %s", cfg.ChainName, cfg.Brand),
		"description":         fmt.Sprintf("Snapshot data, live-peers for %s (%s) by %s", cfg.ChainName, cfg.ChainID, cfg.Brand),
		"chainName":           cfg.ChainName,
		"chainId":             cfg.ChainID,
		"siteName":            fmt.Sprintf("%s validator on %s", cfg.Brand, cfg.ChainName),
		"rpcUrl":              cfg.ExternalResourceRpcUrl,
		"restUrl":             cfg.ExternalResourceRestUrl,
		"grpcUrl":             cfg.ExternalResourceGrpcUrl,
		"logo":                cfg.ExternalResourceLogoUrl,
		"favicon":             cfg.ExternalResourceFaviconUrl,
		"livePeers":           livePeers,
		"livePeersCount":      livePeersCount,
		"generalNodeHomeName": cfg.GeneralNodeHomeName,
		"generalBinaryName":   cfg.GeneralBinaryName,
		"snapshot":            snapshotInfo,
	})
}

func getSnapshotInfo(cfg webtypes.Config) webtypes.SnapshotInfo {
	if ss := cacheSnapshotData.GetRL(); ss != nil {
		return ss.(webtypes.SnapshotInfo)
	}

	ss, _ := cacheSnapshotData.UpdateWL(func() (any, error) {
		ss, err := func() (*webtypes.SnapshotInfo, error) {
			filePath := cfg.SnapshotFilePath
			if filePath == "" {
				return nil, fmt.Errorf("snapshot file path is empty")
			}

			fi, err := os.Stat(filePath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get snapshot file info")
			}

			fileSize := fi.Size()
			if fileSize < 1 {
				return nil, fmt.Errorf("snapshot file is empty")
			}

			var strFileSize string
			if fileSize > 1024*1024*1024 {
				strFileSize = fmt.Sprintf("%.2f GB", float64(fileSize)/1024/1024/1024)
			} else if fileSize > 1024*1024 {
				strFileSize = fmt.Sprintf("%.2f MB", float64(fileSize)/1024/1024)
			} else {
				strFileSize = fmt.Sprintf("%.2f KB", float64(fileSize)/1024)
			}

			var strModTime string
			modTime := time.Since(fi.ModTime())
			if modTime >= 2*24*time.Hour {
				strModTime = fmt.Sprintf("%d days", int(modTime.Hours()/24))
			} else if modTime >= 2*time.Hour {
				strModTime = fmt.Sprintf("%d hours", int(modTime.Hours()))
			} else if modTime >= 2*time.Minute {
				strModTime = fmt.Sprintf("%d minutes", int(modTime.Minutes()))
			} else {
				strModTime = fmt.Sprintf("%d seconds", int(modTime.Seconds()))
			}

			_, fileName := filepath.Split(filePath)
			return &webtypes.SnapshotInfo{
				FileName:         fileName,
				Size:             strFileSize,
				ModTime:          strModTime,
				DownloadFilePath: cfg.SnapshotDownloadURL,
				Error:            nil,
			}, nil
		}()
		if err != nil {
			return webtypes.SnapshotInfo{
				Error: err,
			}, nil
		}
		return *ss, nil
	}, true)

	return ss.(webtypes.SnapshotInfo)
}

func init() {
	cacheSnapshotData = types.NewTimeBasedCache(60 * time.Second)
}
