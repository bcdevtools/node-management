package types

import (
	"sync"
	"time"
)

type TimeBasedCache struct {
	sync.RWMutex
	cache         any
	cacheDuration time.Duration
	expiry        time.Time
}

func NewTimeBasedCache(cacheDuration time.Duration) *TimeBasedCache {
	return &TimeBasedCache{
		cache:         nil,
		cacheDuration: cacheDuration,
		expiry:        time.Time{},
	}
}

func (tbc *TimeBasedCache) GetRL() any {
	return tbc.getRL(true)
}

func (tbc *TimeBasedCache) UpdateWL(funcUpdate func() (any, error), recheckCacheBeforeUpdate bool) (any, error) {
	tbc.Lock()
	defer tbc.Unlock()

	if recheckCacheBeforeUpdate {
		if res := tbc.getRL(false); res != nil {
			return res, nil
		}
	}

	newValue, err := funcUpdate()
	if err != nil {
		return nil, err
	}
	tbc.cache = newValue
	tbc.expiry = time.Now().UTC().Add(tbc.cacheDuration)
	return newValue, nil
}

func (tbc *TimeBasedCache) getRL(acquireLock bool) any {
	if acquireLock {
		tbc.RLock()
		defer tbc.RUnlock()
	}

	if time.Now().UTC().After(tbc.expiry) {
		return nil
	}

	return tbc.cache
}
