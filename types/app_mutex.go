package types

import (
	"encoding/json"
	"fmt"
	"github.com/bcdevtools/node-management/constants"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var mutexAppMutex sync.RWMutex

type AppMutex struct {
	nodeHome           string
	acquiredLock       bool
	aborted            bool
	extendLockDuration time.Duration
}

type lockFileContent struct {
	InstanceID string `json:"instance_id"`
	LockUntil  string `json:"lock_until"`
}

const lockFileName = "." + constants.BINARY_NAME + ".lock"

var instanceIdForLockFile string

func NewAppMutex(nodeHome string, extendLockDuration time.Duration) *AppMutex {
	return &AppMutex{
		nodeHome:           nodeHome,
		acquiredLock:       false,
		aborted:            false,
		extendLockDuration: extendLockDuration,
	}
}

func (am *AppMutex) AcquireLockWL() (bool, error) {
	mutexAppMutex.Lock()
	defer mutexAppMutex.Unlock()

	if am.acquiredLock {
		panic("can not be called twice")
	}

	instanceId, expiry, err := am.readLockUntilRL(false)
	if err != nil {
		return false, errors.Wrap(err, "failed to read lock file")
	}
	if instanceId != "" && instanceId != instanceIdForLockFile {
		return false, fmt.Errorf("lock is already acquired by another instance %s untils %s (%s left)", instanceId, expiry, expiry.Sub(time.Now().UTC()))
	}

	newExpiry := time.Now().UTC().Add(am.extendLockDuration)
	if err := am.writeLockUntilWL(newExpiry, false); err != nil {
		return false, errors.Wrap(err, "failed to write lock file")
	}

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				_, _ = fmt.Fprintln(os.Stderr, "panic in extending lock file:", r)
				os.Exit(1)
			}
		}()

		for {
			time.Sleep(am.extendLockDuration / 4)

			aborted := func() bool {
				mutexAppMutex.RLock()
				defer mutexAppMutex.RUnlock()
				return am.aborted
			}()
			if aborted {
				_ = os.Remove(am.lockFilePath())
				break
			}

			if err := am.writeLockUntilWL(time.Now().UTC().Add(am.extendLockDuration), true); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "failed to extend lock file:", err)
			}
		}
	}()

	am.acquiredLock = true
	return true, nil
}

func (am *AppMutex) writeLockUntilWL(newExpiry time.Time, acquireMutexLock bool) error {
	if acquireMutexLock {
		mutexAppMutex.Lock()
		defer mutexAppMutex.Unlock()
	}

	lfc := lockFileContent{
		InstanceID: instanceIdForLockFile,
		LockUntil:  newExpiry.Format(time.DateTime),
	}

	bz, err := json.Marshal(lfc)
	if err != nil {
		return errors.Wrap(err, "failed to marshal lock file content")
	}

	if err := os.WriteFile(am.lockFilePath(), bz, 0644); err != nil {
		return errors.Wrap(err, "failed to write lock file")
	}

	return nil
}

func (am *AppMutex) readLockUntilRL(acquireMutexLock bool) (string, *time.Time, error) {
	if acquireMutexLock {
		mutexAppMutex.RLock()
		defer mutexAppMutex.RUnlock()
	}

	lockFilePath := am.lockFilePath()
	if _, err := os.Stat(lockFilePath); err != nil {
		if os.IsNotExist(err) {
			return "", nil, nil
		}
		return "", nil, errors.Wrap(err, "failed to check lock file")
	}

	bz, err := os.ReadFile(lockFilePath)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to read lock file")
	}

	var lfc lockFileContent
	if err := json.Unmarshal(bz, &lfc); err != nil {
		return "", nil, errors.Wrap(err, "failed to unmarshal lock file content")
	}

	if lfc.InstanceID == "" || lfc.LockUntil == "" {
		return "", nil, errors.New("lock file is invalid")
	}

	t, err := time.Parse(time.DateTime, lfc.LockUntil)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to parse expiration date from lock file content")
	}

	if t.Before(time.Now().UTC()) {
		return "", nil, nil
	}

	return lfc.InstanceID, &t, nil
}

func (am *AppMutex) ReleaseLockWL() {
	mutexAppMutex.Lock()
	defer mutexAppMutex.Unlock()

	am.aborted = true
}

func (am *AppMutex) lockFilePath() string {
	return filepath.Join(am.nodeHome, lockFileName)
}

func init() {
	nowUTC := time.Now().UTC()
	instanceIdForLockFile = fmt.Sprintf("%s_%d_%s", nowUTC.Format(time.DateTime), nowUTC.UnixNano(), uuid.New().String())
}
