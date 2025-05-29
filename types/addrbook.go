package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net"
	"os"
	"strings"
	"time"
)

type AddrBook struct {
	Key   string          `json:"key"`
	Addrs []*KnownAddress `json:"addrs"`
}

type KnownAddress struct {
	Addr        *NetAddress `json:"addr"`
	Src         *NetAddress `json:"src"`
	Buckets     []int       `json:"buckets"`
	Attempts    int32       `json:"attempts"`
	BucketType  byte        `json:"bucket_type"`
	LastAttempt time.Time   `json:"last_attempt"`
	LastSuccess time.Time   `json:"last_success"`
	LastBanTime time.Time   `json:"last_ban_time"`
}

type NetAddress struct {
	ID   NetAddressID `json:"id"`
	IP   net.IP       `json:"ip"`
	Port uint16       `json:"port"`
}

type NetAddressID string

func (ab *AddrBook) ReadAddrBook(inputFilePath string) error {
	bz, err := os.ReadFile(inputFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	err = json.Unmarshal(bz, ab)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}

	return nil
}

func (ab *AddrBook) GetLivePeers(validDuration time.Duration, excludeIPv6 bool) []*KnownAddress {
	var livePeers []*KnownAddress
	for _, addr := range ab.Addrs {
		if addr.Addr == nil {
			continue
		}

		if addr.LastSuccess.IsZero() || addr.LastAttempt.IsZero() { // means not any success
			continue
		}

		if addr.LastSuccess.Before(addr.LastAttempt) { // means not connected atm
			if addr.LastAttempt.Sub(addr.LastSuccess) > validDuration {
				continue
			}
			if time.Since(addr.LastAttempt) > validDuration {
				continue
			}
		}

		if excludeIPv6 {
			spl := strings.Split(addr.Addr.IP.String(), ":")
			if len(spl) > 2 {
				continue
			}
		}

		livePeers = append(livePeers, addr)
	}

	return livePeers
}
