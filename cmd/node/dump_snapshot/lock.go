package dump_snapshot

import (
	"fmt"
	"github.com/bcdevtools/node-management/types"
	"github.com/pkg/errors"
	"time"
)

func acquireSingletonInstance(nodeHomeDirectory, dumpHomeDir string) (appOriginalNodeMutex, appDumpNodeMutex *types.AppMutex, err error) {
	appOriginalNodeMutex = types.NewAppMutex(nodeHomeDirectory, 4*time.Second)
	if acquiredLock, errAcquire := appOriginalNodeMutex.AcquireLockWL(); errAcquire != nil {
		err = errors.Wrap(errAcquire, "failed to acquire lock single instance in original node home")
		return
	} else if !acquiredLock {
		err = fmt.Errorf("failed to acquire lock single instance in original node")
		return
	}

	appDumpNodeMutex = types.NewAppMutex(dumpHomeDir, 8*time.Second)
	if acquiredLock, errAcquire := appDumpNodeMutex.AcquireLockWL(); errAcquire != nil {
		err = errors.Wrap(errAcquire, "failed to acquire lock single instance in dump node")
		return
	} else if !acquiredLock {
		err = fmt.Errorf("failed to acquire lock single instance in dump node")
		return
	}

	return
}
