package chetcd

import (
	"fmt"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"strings"
	"time"
)

type DirWatcher struct {
	keysAPI client.KeysAPI

	dir          string
	watcher      client.Watcher
	keyObservers map[string]*dirWatcherObserver
}

func NewDirWatcher(keysAPI client.KeysAPI, dir string) *DirWatcher {
	return &DirWatcher{
		keysAPI: keysAPI,
		dir:     dir,
	}
}

func (d *DirWatcher) fullKeyDir(keyRelativeToDir string) string {
	return strings.TrimRight(d.dir, "/") + "/" + strings.TrimLeft(keyRelativeToDir, "/")
}

func (d *DirWatcher) putObserver(configValue *ConfigValue) <-chan *ValueChange {
	if d.keyObservers == nil {
		d.keyObservers = make(map[string]*dirWatcherObserver)
	}

	valueChanges := make(chan *ValueChange)
	d.keyObservers[configValue.fullKey] = &dirWatcherObserver{configValue.modifiedIndex, valueChanges}
	return valueChanges
}

func (d *DirWatcher) GetObservedConfigValue(keyRelativeToDir string) (*ConfigValue, error) {
	fullKey := d.fullKeyDir(keyRelativeToDir)

	resp, err := d.keysAPI.Get(context.Background(), fullKey, nil)
	if err != nil {
		return nil, err
	}

	if resp.Node.Key != fullKey {
		return nil, fmt.Errorf("Unexpected error, keys do not match: '%s' != '%s'", resp.Node.Key, fullKey)
	}

	c := &ConfigValue{
		keysAPI:       d.keysAPI,
		fullKey:       fullKey,
		modifiedIndex: resp.Node.ModifiedIndex,
		value:         resp.Node.Value,
	}

	valueChangesChannel := d.putObserver(c)
	c.valueChangesChannel = valueChangesChannel

	return c, nil
}

func (d *DirWatcher) WatchDir(delayOnError time.Duration) <-chan error {
	errs := make(chan error)
	go d.handleWatcherChannels(delayOnError, errs)
	return errs
}

func (d *DirWatcher) handleWatcherChannels(delayOnError time.Duration, errs chan<- error) {
	defer func() {
		close(errs)
		for _, obs := range d.keyObservers {
			close(obs.valueChanges)
		}
	}()

	d.watcher = d.keysAPI.Watcher(d.dir, &client.WatcherOptions{Recursive: true, AfterIndex: 1})

	for {
		resp, err := d.watcher.Next(context.Background())
		if err != nil {
			errMsg := fmt.Errorf("ERROR in watcher.Next(): %s", err.Error())
			fmt.Printf("%s\n", errMsg.Error())
			errs <- errMsg
			time.Sleep(delayOnError)
			continue
		}

		if obs, ok := d.keyObservers[resp.Node.Key]; ok {
			if resp.Node.ModifiedIndex > obs.afterModifiedIndex {
				obs.valueChanges <- &ValueChange{NewNode: *resp.Node, OldNode: *resp.Node}
			}
		}
	}
}
