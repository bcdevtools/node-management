package types

import "path"

type Config struct {
	Port           uint16
	AuthorizeToken string
	NodeHome       string
	Debug          bool
}

func (c Config) GetAddrBookFilePath() string {
	return path.Join(c.NodeHome, "config", "addrbook.json")
}
