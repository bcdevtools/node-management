package types

import "path"

type Config struct {
	Port           uint16
	AuthorizeToken string
	NodeHome       string
	Debug          bool

	Brand string

	// General chain-based configuration
	ChainName           string // Evmos Dymension etc
	ChainID             string // evmos_9001-2 dymension_1100-1 etc
	ChainDescription    string
	GeneralBinaryName   string // evmosd dymd etc
	GeneralNodeHomeName string // .evmosd .dymension etc

	// External web resources
	ExternalResourceLogoUrl    string
	ExternalResourceFaviconUrl string
	ExternalResourceRpcUrl     string
	ExternalResourceRestUrl    string
	ExternalResourceGrpcUrl    string

	// Snapshot information
	SnapshotFilePath    string
	SnapshotDownloadURL string
}

func (c Config) GetAddrBookFilePath() string {
	return path.Join(c.NodeHome, "config", "addrbook.json")
}
