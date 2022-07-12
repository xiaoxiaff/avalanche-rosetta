package indexer

import (
	"sync"
	"time"
)

type ChainTime struct {
	sync.RWMutex
	value time.Time

	proposedValue time.Time
}

func (t *ChainTime) Read() int64 {
	t.RLock()
	defer t.RUnlock()
	return t.value.Unix()
}

func (t *ChainTime) Write(value time.Time) {
	t.Lock()
	t.value = value
	t.Unlock()
}

func (t *ChainTime) ProposeWrite(value time.Time) {
	t.proposedValue = value
}

func (t *ChainTime) ReadProposedWrite() int64 {
	return t.proposedValue.Unix()
}

func (t *ChainTime) AcceptProposedWrite() {
	if t.proposedValue.IsZero() {
		return
	}

	t.Write(t.proposedValue)
}

func (t *ChainTime) RejectProposedWrite() {
	t.proposedValue = time.Unix(0, 0)
}
