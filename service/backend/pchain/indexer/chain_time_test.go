package indexer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	t0 = time.Unix(0, 0)
	t1 = time.Unix(100, 0)
	t2 = time.Unix(200, 0)
	t3 = time.Unix(300, 0)
)

func TestChainTime(t *testing.T) {
	var ti ChainTime

	assert := assert.New(t)

	// Should update [value] without updating [proposedValue]
	ti.Write(t1)
	assert.NotEqual(ti.proposedValue, t1)
	assert.NotEqual(ti.ReadProposedWrite(), t1.Unix())
	assert.Equal(ti.value, t1)
	assert.Equal(ti.Read(), t1.Unix())

	// Should set [proposedValue] to [t3]
	// [value] should be unchanged
	ti.ProposeWrite(t2)
	assert.Equal(ti.proposedValue, t2)
	assert.Equal(ti.ReadProposedWrite(), t2.Unix())
	assert.Equal(ti.value, t1)
	assert.Equal(ti.Read(), t1.Unix())

	// Should set [proposedValue] to [time.Unix(0, 0)]
	// [value] should be unchanged
	ti.RejectProposedWrite()
	assert.Equal(ti.proposedValue, t0)
	assert.Equal(ti.ReadProposedWrite(), t0.Unix())
	assert.Equal(ti.value, t1)
	assert.Equal(ti.Read(), t1.Unix())

	// Should set [proposedValue] to [t3]
	// [value] should be unchanged
	ti.ProposeWrite(t3)
	assert.Equal(ti.proposedValue, t3)
	assert.Equal(ti.ReadProposedWrite(), t3.Unix())
	assert.Equal(ti.value, t1)
	assert.Equal(ti.Read(), t1.Unix())

	// Should set [value] to [proposedValue]
	// [proposedValue] should be unchanged
	ti.AcceptProposedWrite()
	assert.Equal(ti.proposedValue, ti.value)
	assert.Equal(ti.ReadProposedWrite(), ti.value.Unix())
	assert.Equal(ti.value, t3)
	assert.Equal(ti.Read(), t3.Unix())

	// All values should be unchanged on a second
	// [AcceptProposedWrite] call
	ti.AcceptProposedWrite()
	assert.Equal(ti.proposedValue, ti.value)
	assert.Equal(ti.ReadProposedWrite(), ti.value.Unix())
	assert.Equal(ti.value, t3)
	assert.Equal(ti.Read(), t3.Unix())
}
