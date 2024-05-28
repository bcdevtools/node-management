package dump_snapshot

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"path"
	"strings"
)

func getServiceName(noService bool, binary string, cmd *cobra.Command) (serviceName string, err error) {
	if noService {
		return
	}

	defer func() {
		if err != nil {
			serviceName = ""
		}
	}()

	customServiceName, _ := cmd.Flags().GetString(flagServiceName)
	if customServiceName != "" {
		if strings.Contains(customServiceName, "/") {
			err = fmt.Errorf("service name cannot contain path, provide name only")
			return
		}
		serviceName = customServiceName
	} else {
		_, binaryName := path.Split(binary)
		if binaryName == "" {
			err = fmt.Errorf("failed to get service name from binary path, require flag --%s\n", flagServiceName)
			return
		}
		serviceName = binaryName
	}

	serviceName = strings.TrimSuffix(serviceName, ".service")

	expectedServiceFile := path.Join("/etc/systemd/system", serviceName+".service")
	_, exists, _, errChkSvcF := utils.FileInfo(expectedServiceFile)
	if errChkSvcF != nil {
		err = errors.Wrapf(errChkSvcF, "failed to check service file %s", expectedServiceFile)
		return
	}
	if !exists {
		err = fmt.Errorf("expected service file does not exists [%s], correct service file name by flag --%s", expectedServiceFile, flagServiceName)
		return
	}

	return
}
