package validation

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/pkg/errors"
	"strings"
)

func ValidateNodeBinary(binary string) error {
	if binary == "" {
		return fmt.Errorf("binary cannot be empty")
	}

	if strings.Contains(binary, "/") {
		_, exists, isDir, err := utils.FileInfo(binary)
		if err != nil {
			return errors.Wrap(err, "failed to check binary path")
		}
		if !exists {
			return fmt.Errorf("specified binary does not exist: %s", binary)
		}
		if isDir {
			return fmt.Errorf("specified binary is a directory: %s", binary)
		}

		return nil
	}

	if !utils.HasBinaryName(binary) {
		return fmt.Errorf("binary name %s might not available in $PATH", binary)
	}

	return nil
}
