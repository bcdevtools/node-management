package types

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type P2pConfigToml struct {
	Seeds               string `toml:"seeds"`
	Laddr               string `toml:"laddr"`
	PersistentPeers     string `toml:"persistent_peers"`
	MaxNumInboundPeers  int    `toml:"max_num_inbound_peers"`
	MaxNumOutboundPeers int    `toml:"max_num_outbound_peers"`
	SeedMode            bool   `toml:"seed_mode"`
}

type StateSyncConfigToml struct {
	Enable bool `toml:"enable"`
}

type ConsensusConfigToml struct {
	DoubleSignCheckHeight uint `toml:"double_sign_check_height"`
	SkipTimeoutCommit     bool `toml:"skip_timeout_commit"`
}

type TxIndexConfigToml struct {
	Indexer string `toml:"indexer"`
}

type RpcConfigToml struct {
	LAddr string `toml:"laddr"`
}

type ConfigToml struct {
	Moniker   string               `toml:"moniker"`
	P2P       *P2pConfigToml       `toml:"p2p"`
	StateSync *StateSyncConfigToml `toml:"statesync"`
	Consensus *ConsensusConfigToml `toml:"consensus"`
	TxIndex   *TxIndexConfigToml   `toml:"tx_index"`
	RPC       *RpcConfigToml       `toml:"rpc"`
}

func ReadNodeRpcFromConfigToml(configFilePath string) (rpc string, err error) {
	var exists bool
	_, exists, _, err = utils.FileInfo(configFilePath)
	if err != nil {
		err = errors.Wrap(err, "failed to check "+configFilePath)
		return
	}
	if !exists {
		err = fmt.Errorf("file not found: " + configFilePath)
		return
	}

	var bz []byte
	bz, err = os.ReadFile(configFilePath)
	if err != nil {
		err = errors.Wrap(err, "failed to read "+configFilePath)
		return
	}

	var config ConfigToml
	err = toml.Unmarshal(bz, &config)
	if err != nil {
		err = errors.Wrap(err, "failed to unmarshal "+configFilePath)
		return
	}
	if config.RPC == nil || config.RPC.LAddr == "" {
		err = fmt.Errorf("rpc section, address is not set in " + configFilePath)
		return
	}

	addr := strings.TrimSpace(config.RPC.LAddr)
	addr = strings.TrimPrefix(addr, "tcp://")
	addr = strings.TrimSuffix(addr, "/")
	//goland:noinspection HttpUrlsUsage
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}

	return addr, nil
}
