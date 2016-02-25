package chetcd

import (
	"github.com/coreos/etcd/client"
)

type ValueChange struct {
	OldNode client.Node
	NewNode client.Node
}
