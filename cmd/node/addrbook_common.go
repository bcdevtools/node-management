package node

import (
	"encoding/json"
	"github.com/bcdevtools/node-management/types"
	"github.com/pkg/errors"
	"os"
	"time"
)

func readAddrBook(inputFilePath string) (*types.AddrBook, error) {
	bz, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var addrBook types.AddrBook
	err = json.Unmarshal(bz, &addrBook)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}

	return &addrBook, nil
}

func getLivePeers(addrBook *types.AddrBook, validDuration time.Duration) []*types.KnownAddress {
	var livePeers []*types.KnownAddress
	for _, addr := range addrBook.Addrs {
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

		livePeers = append(livePeers, addr)
	}

	return livePeers
}
