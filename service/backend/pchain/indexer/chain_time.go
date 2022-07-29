package indexer

import (
	"sync"
	"time"
)

type ChainTime struct {
	valueMutex sync.RWMutex
	value      time.Time

	proposedValueMutex sync.RWMutex
	proposedValue      time.Time
}

func (t *ChainTime) Read() int64 {
	t.valueMutex.RLock()
	defer t.valueMutex.RUnlock()
	return t.value.Unix()
}

func (t *ChainTime) Write(value time.Time) {
	t.valueMutex.Lock()
	t.value = value
	t.valueMutex.Unlock()
}

func (t *ChainTime) ProposeWrite(value time.Time) {
	t.proposedValueMutex.Lock()
	t.proposedValue = value
	t.proposedValueMutex.Unlock()
}

func (t *ChainTime) ReadProposedWrite() int64 {
	t.proposedValueMutex.RLock()
	defer t.proposedValueMutex.RUnlock()
	return t.proposedValue.Unix()
}

func (t *ChainTime) AcceptProposedWrite() {
	t.proposedValueMutex.RLock()
	defer t.proposedValueMutex.RUnlock()
	if t.proposedValue.IsZero() || t.proposedValue.Unix() == 0 {
		return
	}

	t.Write(t.proposedValue)
}

func (t *ChainTime) RejectProposedWrite() {
	t.proposedValueMutex.Lock()
	t.proposedValue = time.Unix(0, 0)
	t.proposedValueMutex.Unlock()
}
