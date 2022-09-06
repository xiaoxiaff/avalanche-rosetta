// Code generated by mockery v2.12.3. DO NOT EDIT.

package chain

import (
	common "github.com/ava-labs/avalanche-rosetta/service/backend/common"
	mock "github.com/stretchr/testify/mock"

	types "github.com/coinbase/rosetta-sdk-go/types"
)

// TxCombiner is an autogenerated mock type for the TxCombiner type
type TxCombiner struct {
	mock.Mock
}

// CombineTx provides a mock function with given fields: tx, signatures
func (_m *TxCombiner) CombineTx(tx common.AvaxTx, signatures []*types.Signature) (common.AvaxTx, *types.Error) {
	ret := _m.Called(tx, signatures)

	var r0 common.AvaxTx
	if rf, ok := ret.Get(0).(func(common.AvaxTx, []*types.Signature) common.AvaxTx); ok {
		r0 = rf(tx, signatures)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.AvaxTx)
		}
	}

	var r1 *types.Error
	if rf, ok := ret.Get(1).(func(common.AvaxTx, []*types.Signature) *types.Error); ok {
		r1 = rf(tx, signatures)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*types.Error)
		}
	}

	return r0, r1
}

type NewTxCombinerT interface {
	mock.TestingT
	Cleanup(func())
}

// NewTxCombiner creates a new instance of TxCombiner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTxCombiner(t NewTxCombinerT) *TxCombiner {
	mock := &TxCombiner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
