package chetcd

type dirWatcherObserver struct {
	afterModifiedIndex uint64
	valueChanges       chan<- *ValueChange
}
