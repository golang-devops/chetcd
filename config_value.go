package chetcd

import (
	"github.com/coreos/etcd/client"
)

type ConfigValue struct {
	keysAPI       client.KeysAPI
	fullKey       string
	modifiedIndex uint64

	value               string
	valueChangesChannel <-chan *ValueChange
}

func (c *ConfigValue) FullKey() string                     { return c.fullKey }
func (c *ConfigValue) Value() string                       { return c.value }
func (c *ConfigValue) String() string                      { return c.value }
func (c *ConfigValue) ModifiedIndex() uint64               { return c.modifiedIndex }
func (c *ConfigValue) ChangesChannel() <-chan *ValueChange { return c.valueChangesChannel }
