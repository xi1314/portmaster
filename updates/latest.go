package updates

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Safing/portbase/log"
)

var (
	stableUpdates = make(map[string]string)
	betaUpdates   = make(map[string]string)
	localUpdates  = make(map[string]string)
	updatesLock   sync.RWMutex
)

// ReloadLatest reloads available updates from disk.
func ReloadLatest() error {
	newLocalUpdates := make(map[string]string)

	// all
	prefix := "all"
	new, err1 := ScanForLatest(filepath.Join(updateStoragePath, prefix), false)
	for key, val := range new {
		newLocalUpdates[filepath.Join(prefix, key)] = val
	}

	// os_platform
	prefix = fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	new, err2 := ScanForLatest(filepath.Join(updateStoragePath, prefix), false)
	for key, val := range new {
		newLocalUpdates[filepath.Join(prefix, key)] = val
	}

	if err1 != nil && err2 != nil {
		return fmt.Errorf("could not load latest update versions: %s, %s", err1, err2)
	}

	log.Tracef("updates: loading latest updates:")

	for key, val := range newLocalUpdates {
		log.Tracef("updates: %s v%s", key, val)
	}

	updatesLock.Lock()
	localUpdates = newLocalUpdates
	updatesLock.Unlock()

	log.Tracef("updates: load complete")

	if len(stableUpdates) == 0 {
		err := loadIndexesFromDisk()
		if err != nil {
			return err
		}
	}

	// update version status
	updatesLock.RLock()
	defer updatesLock.RUnlock()
	updateStatus(versionClassLocal, localUpdates)

	return nil
}

func ScanForLatest(baseDir string, hardFail bool) (latest map[string]string, lastError error) {
	var added int
	latest = make(map[string]string)

	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			lastError = err
			if hardFail {
				return err
			}
			log.Warningf("updates: could not read %s", path)
			return nil
		}
		if !info.IsDir() {
			added++
		}

		relativePath := strings.TrimLeft(strings.TrimPrefix(path, baseDir), "/")
		identifierPath, version, ok := getIdentifierAndVersion(relativePath)
		if !ok {
			return nil
		}

		// add/update index
		storedVersion, ok := latest[identifierPath]
		if ok {
			// FIXME: this will fail on multi-digit version segments!
			if version > storedVersion {
				latest[identifierPath] = version
			}
		} else {
			latest[identifierPath] = version
		}

		return nil
	})

	if lastError != nil {
		if hardFail {
			return nil, lastError
		}
		if added == 0 {
			return latest, lastError
		}
	}
	return latest, nil
}

func loadIndexesFromDisk() error {
	data, err := ioutil.ReadFile(filepath.Join(updateStoragePath, "stable.json"))
	if err != nil {
		if os.IsNotExist(err) {
			log.Infof("updates: stable.json does not yet exist, waiting for first update cycle")
			return nil
		}
		return err
	}

	newStableUpdates := make(map[string]string)
	err = json.Unmarshal(data, &newStableUpdates)
	if err != nil {
		return err
	}

	if len(newStableUpdates) == 0 {
		return errors.New("stable.json is empty")
	}

	log.Tracef("updates: loaded stable.json")

	updatesLock.Lock()
	stableUpdates = newStableUpdates
	updatesLock.Unlock()

	// update version status
	updatesLock.RLock()
	defer updatesLock.RUnlock()
	updateStatus(versionClassStable, stableUpdates)

	return nil
}
