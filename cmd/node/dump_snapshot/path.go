package dump_snapshot

import (
	"fmt"
	"github.com/bcdevtools/node-management/types"
	"github.com/bcdevtools/node-management/utils"
	"github.com/pkg/errors"
	"os"
	"path"
)

func prepareDumpNodeHomeDirectory(dumpHomeDir, nodeHomeDir string) error {
	_, exists, _, err := utils.FileInfo(dumpHomeDir)
	if err != nil {
		return errors.Wrap(err, "failed to check dump home directory at "+dumpHomeDir)
	}

	if !exists {
		fmt.Println("INF: creating dump home directory")
		err = os.Mkdir(dumpHomeDir, 0o755)
		if err != nil {
			return errors.Wrap(err, "failed to create dump home directory at "+dumpHomeDir)
		}
	}

	configDirOfDump := path.Join(dumpHomeDir, "config")
	err = os.Mkdir(configDirOfDump, 0o755)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to create config directory at "+configDirOfDump)
	}

	dataDirOfDump := path.Join(dumpHomeDir, "data")
	err = os.Mkdir(dataDirOfDump, 0o700)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "failed to create data directory at "+dataDirOfDump)
	}

	privValStateJsonFilePath := path.Join(dataDirOfDump, "priv_validator_state.json")
	err = (&types.PrivateValidatorState{
		Height: "0",
	}).SaveToJSONFile(privValStateJsonFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to create empty "+privValStateJsonFilePath)
	}

	copyConfig := func(fileName string) error {
		src := path.Join(nodeHomeDir, "config", fileName)
		dst := path.Join(dumpHomeDir, "config", fileName)
		ec := utils.LaunchApp("cp", []string{src, dst})
		if ec != 0 {
			return errors.Wrap(err, "failed to copy config file "+fileName)
		}
		fmt.Println("INF: copied config file:", fileName)
		return nil
	}

	fmt.Println("INF: Copying config files")
	if err = copyConfig("config.toml"); err != nil {
		return err
	}
	if err = copyConfig("app.toml"); err != nil {
		return err
	}
	if err = copyConfig("genesis.json"); err != nil {
		return err
	}
	if err = copyConfig("client.toml"); err != nil {
		return err
	}
	return nil
}
