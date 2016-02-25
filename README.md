# chetcd
Config-helpers for etcd. The aim is to have some common etcd-client use cases wrapper in an elegant API.


## Dir watcher

A dir watcher will simply watch a (directory) key recursively. We can attach key observers to receive these changes via native golang channels.

A typical usage would be if we have a directory for all our config values of the application. We want to watch all those config values and react on them to update our live instance.


### Quick Start


#### Source

`go get -u github.com/coreos/etcd`
`go get -u github.com/francoishill/chetcd`

#### Usage

```
package main

import (
    "fmt"
    "github.com/coreos/etcd/client"
    "github.com/francoishill/chetcd"
    "log"
    "time"
)

func main() {
    cfg := client.Config{
        Endpoints:               []string{"http://127.0.0.1:2379"},
        Transport:               client.DefaultTransport,
        HeaderTimeoutPerRequest: time.Second, // set timeout per request to fail fast when the target endpoint is unavailable
    }

    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    keysAPI := client.NewKeysAPI(c)

    foodirWatcher := chetcd.NewDirWatcher(keysAPI, "/foodir")

    bar1ConfigValue, err := foodirWatcher.GetObservedConfigValue("/bar1")
    if err != nil {
        log.Fatal(err)
    }

    bar2ConfigValue, err := foodirWatcher.GetObservedConfigValue("/bar2")
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        for {
            select {
            case b1 := <-bar1ConfigValue.ChangesChannel():
                fmt.Printf("Bar 1 changed from '%s' to '%s'\n", b1.OldNode.Value, b1.NewNode.Value)
                break
            case b2 := <-bar2ConfigValue.ChangesChannel():
                fmt.Printf("Bar 2 changed from '%s' to '%s'\n", b2.OldNode.Value, b2.NewNode.Value)
                break
            }
        }
    }()

    //The delay to wait when we receive `watcher.Next()` errors. This could be because the etcd server is down.
    delayOnError := 30 * time.Second
    errChan := foodirWatcher.WatchDir(delayOnError)
    for er := range errChan {
        fmt.Printf("DIR WATCH ERROR: %s, watcher: %q\n", er.Error(), foodirWatcher)
    }
}
```