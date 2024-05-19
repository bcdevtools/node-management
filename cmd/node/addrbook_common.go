package node

import (
	"github.com/bcdevtools/node-management/types"
	"github.com/pkg/errors"
)

func readAddrBook(inputFilePath string) (*types.AddrBook, error) {
	addrBook := &types.AddrBook{}
	if err := addrBook.ReadAddrBook(inputFilePath); err != nil {
		return nil, errors.Wrap(err, "failed to read addrbook")
	}
	return addrBook, nil
}
