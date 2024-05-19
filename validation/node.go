package validation

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/pkg/errors"
	"path"
	"regexp"
)

var regexPeerPlus = regexp.MustCompile(`^[a-f\d]{40}@(([^:]+)|(\[[a-f\d]*(:+[a-f\d]+)+])):\d{1,5}(,[a-f\d]{40}@(([^:]+)|(\[[a-f\d]*(:+[a-f\d]+)+])):\d{1,5})*$`)

func IsValidPeer(peer string) bool {
	return regexPeerPlus.MatchString(peer)
}

func PossibleNodeHome(nodeHomeDirectory string) error {
	if nodeHomeDirectory == "" {
		return fmt.Errorf("node home directory cannot be empty")
	}

	_, exists, isDir, err := utils.FileInfo(nodeHomeDirectory)
	if err != nil {
		return errors.Wrap(err, "failed to check node home directory")
	}
	if !exists {
		return fmt.Errorf("specified node home directory does not exist: %s", nodeHomeDirectory)
	}
	if !isDir {
		return fmt.Errorf("specified path of node home directory is not a directory: %s", nodeHomeDirectory)
	}

	configDirPath := path.Join(nodeHomeDirectory, "config")
	_, exists, isDir, err = utils.FileInfo(configDirPath)
	if err != nil {
		return errors.Wrap(err, "failed to check node config directory")
	}
	if !exists {
		return fmt.Errorf("node config directory does not exist: %s", configDirPath)
	}
	if !isDir {
		return fmt.Errorf("specified path of node config directory is not a directory: %s", configDirPath)
	}

	dataDirPath := path.Join(nodeHomeDirectory, "data")
	_, exists, isDir, err = utils.FileInfo(dataDirPath)
	if err != nil {
		return errors.Wrap(err, "failed to check node data directory")
	}
	if !exists {
		return fmt.Errorf("node data directory does not exist: %s", dataDirPath)
	}
	if !isDir {
		return fmt.Errorf("specified path of node data directory is not a directory: %s", dataDirPath)
	}

	return nil
}
