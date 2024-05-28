package dump_snapshot

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/bcdevtools/node-management/validation"
	"github.com/pkg/errors"
	"path"
	"strings"
)

func validateBinary(binary string) error {
	if binary == "" {
		return fmt.Errorf("binary path is required")
	}

	if !strings.Contains(binary, "/") {
		if !utils.HasBinaryName(binary) {
			return fmt.Errorf("specified binary does not exists or not included in PATH: %s", binary)
		}

		return nil
	}

	_, exists, isDir, err := utils.FileInfo(binary)
	if err != nil {
		return errors.Wrapf(err, "failed to check binary path: %s", binary)
	}
	if !exists {
		return fmt.Errorf("specified binary does not exists: %s", binary)
	}
	if isDir {
		return fmt.Errorf("specified binary path is a directory: %s", binary)
	}
	return nil
}

func validateNodeHomeDirectory(nodeHomeDirectory string) error {
	err := validation.PossibleNodeHome(nodeHomeDirectory)
	if err != nil {
		return errors.Wrapf(err, "invalid node home directory: %s", nodeHomeDirectory)
	}
	_, exists, _, err := utils.FileInfo(path.Join(nodeHomeDirectory, "data", "application.db"))
	if err != nil {
		return errors.Wrapf(err, "failed to check application.db in node home directory: %s", nodeHomeDirectory)
	}
	if !exists {
		return fmt.Errorf("node home directory does not contains data")
	}
	return nil
}

func validateOutputFile(outputFile string) error {
	if outputFile == "" {
		return fmt.Errorf("output file path is required")
	}
	_, exists, _, err := utils.FileInfo(outputFile)
	if err != nil {
		return errors.Wrapf(err, "failed to check output file path %s", outputFile)
	}
	if !exists {
		return fmt.Errorf("output file path does not exists: %s", outputFile)
	}
	if !strings.HasSuffix(outputFile, ".tar.gz") {
		return fmt.Errorf("output file must be .tar.gz: %s", outputFile)
	}
	return nil
}
