// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package client

import (
	api "github.com/ava-labs/avalanchego/api"
	avm "github.com/ava-labs/avalanchego/vms/avm"

	context "context"

	ids "github.com/ava-labs/avalanchego/ids"

	indexer "github.com/ava-labs/avalanchego/indexer"

	info "github.com/ava-labs/avalanchego/api/info"

	mock "github.com/stretchr/testify/mock"

	platformvm "github.com/ava-labs/avalanchego/vms/platformvm"

	rpc "github.com/ava-labs/avalanchego/utils/rpc"
)

// PChainClient is an autogenerated mock type for the PChainClient type
type PChainClient struct {
	mock.Mock
}

// GetAssetDescription provides a mock function with given fields: ctx, assetID, options
func (_m *PChainClient) GetAssetDescription(ctx context.Context, assetID string, options ...rpc.Option) (*avm.GetAssetDescriptionReply, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, assetID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *avm.GetAssetDescriptionReply
	if rf, ok := ret.Get(0).(func(context.Context, string, ...rpc.Option) *avm.GetAssetDescriptionReply); ok {
		r0 = rf(ctx, assetID, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*avm.GetAssetDescriptionReply)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, ...rpc.Option) error); ok {
		r1 = rf(ctx, assetID, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAtomicUTXOs provides a mock function with given fields: ctx, addrs, sourceChain, limit, startAddress, startUTXOID, options
func (_m *PChainClient) GetAtomicUTXOs(ctx context.Context, addrs []ids.ShortID, sourceChain string, limit uint32, startAddress ids.ShortID, startUTXOID ids.ID, options ...rpc.Option) ([][]byte, ids.ShortID, ids.ID, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, addrs, sourceChain, limit, startAddress, startUTXOID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func(context.Context, []ids.ShortID, string, uint32, ids.ShortID, ids.ID, ...rpc.Option) [][]byte); ok {
		r0 = rf(ctx, addrs, sourceChain, limit, startAddress, startUTXOID, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 ids.ShortID
	if rf, ok := ret.Get(1).(func(context.Context, []ids.ShortID, string, uint32, ids.ShortID, ids.ID, ...rpc.Option) ids.ShortID); ok {
		r1 = rf(ctx, addrs, sourceChain, limit, startAddress, startUTXOID, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(ids.ShortID)
		}
	}

	var r2 ids.ID
	if rf, ok := ret.Get(2).(func(context.Context, []ids.ShortID, string, uint32, ids.ShortID, ids.ID, ...rpc.Option) ids.ID); ok {
		r2 = rf(ctx, addrs, sourceChain, limit, startAddress, startUTXOID, options...)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(ids.ID)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(context.Context, []ids.ShortID, string, uint32, ids.ShortID, ids.ID, ...rpc.Option) error); ok {
		r3 = rf(ctx, addrs, sourceChain, limit, startAddress, startUTXOID, options...)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetBalance provides a mock function with given fields: ctx, addrs, options
func (_m *PChainClient) GetBalance(ctx context.Context, addrs []ids.ShortID, options ...rpc.Option) (*platformvm.GetBalanceResponse, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, addrs)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *platformvm.GetBalanceResponse
	if rf, ok := ret.Get(0).(func(context.Context, []ids.ShortID, ...rpc.Option) *platformvm.GetBalanceResponse); ok {
		r0 = rf(ctx, addrs, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*platformvm.GetBalanceResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []ids.ShortID, ...rpc.Option) error); ok {
		r1 = rf(ctx, addrs, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBlock provides a mock function with given fields: ctx, blockID, options
func (_m *PChainClient) GetBlock(ctx context.Context, blockID ids.ID, options ...rpc.Option) ([]byte, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, blockID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, ids.ID, ...rpc.Option) []byte); ok {
		r0 = rf(ctx, blockID, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ids.ID, ...rpc.Option) error); ok {
		r1 = rf(ctx, blockID, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBlockchainID provides a mock function with given fields: _a0, _a1, _a2
func (_m *PChainClient) GetBlockchainID(_a0 context.Context, _a1 string, _a2 ...rpc.Option) (ids.ID, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 ids.ID
	if rf, ok := ret.Get(0).(func(context.Context, string, ...rpc.Option) ids.ID); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ids.ID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContainerByID provides a mock function with given fields: ctx, containerID, options
func (_m *PChainClient) GetContainerByID(ctx context.Context, containerID ids.ID, options ...rpc.Option) (indexer.Container, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, containerID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 indexer.Container
	if rf, ok := ret.Get(0).(func(context.Context, ids.ID, ...rpc.Option) indexer.Container); ok {
		r0 = rf(ctx, containerID, options...)
	} else {
		r0 = ret.Get(0).(indexer.Container)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ids.ID, ...rpc.Option) error); ok {
		r1 = rf(ctx, containerID, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContainerByIndex provides a mock function with given fields: ctx, index, options
func (_m *PChainClient) GetContainerByIndex(ctx context.Context, index uint64, options ...rpc.Option) (indexer.Container, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, index)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 indexer.Container
	if rf, ok := ret.Get(0).(func(context.Context, uint64, ...rpc.Option) indexer.Container); ok {
		r0 = rf(ctx, index, options...)
	} else {
		r0 = ret.Get(0).(indexer.Container)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint64, ...rpc.Option) error); ok {
		r1 = rf(ctx, index, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetHeight provides a mock function with given fields: ctx, options
func (_m *PChainClient) GetHeight(ctx context.Context, options ...rpc.Option) (uint64, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(context.Context, ...rpc.Option) uint64); ok {
		r0 = rf(ctx, options...)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...rpc.Option) error); ok {
		r1 = rf(ctx, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLastAccepted provides a mock function with given fields: _a0, _a1
func (_m *PChainClient) GetLastAccepted(_a0 context.Context, _a1 ...rpc.Option) (indexer.Container, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 indexer.Container
	if rf, ok := ret.Get(0).(func(context.Context, ...rpc.Option) indexer.Container); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Get(0).(indexer.Container)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNetworkID provides a mock function with given fields: _a0, _a1
func (_m *PChainClient) GetNetworkID(_a0 context.Context, _a1 ...rpc.Option) (uint32, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 uint32
	if rf, ok := ret.Get(0).(func(context.Context, ...rpc.Option) uint32); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Get(0).(uint32)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNodeID provides a mock function with given fields: _a0, _a1
func (_m *PChainClient) GetNodeID(_a0 context.Context, _a1 ...rpc.Option) (ids.NodeID, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 ids.NodeID
	if rf, ok := ret.Get(0).(func(context.Context, ...rpc.Option) ids.NodeID); ok {
		r0 = rf(_a0, _a1...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ids.NodeID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRewardUTXOs provides a mock function with given fields: _a0, _a1, _a2
func (_m *PChainClient) GetRewardUTXOs(_a0 context.Context, _a1 *api.GetTxArgs, _a2 ...rpc.Option) ([][]byte, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func(context.Context, *api.GetTxArgs, ...rpc.Option) [][]byte); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *api.GetTxArgs, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStake provides a mock function with given fields: ctx, addrs, options
func (_m *PChainClient) GetStake(ctx context.Context, addrs []ids.ShortID, options ...rpc.Option) (uint64, [][]byte, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, addrs)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(context.Context, []ids.ShortID, ...rpc.Option) uint64); ok {
		r0 = rf(ctx, addrs, options...)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 [][]byte
	if rf, ok := ret.Get(1).(func(context.Context, []ids.ShortID, ...rpc.Option) [][]byte); ok {
		r1 = rf(ctx, addrs, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([][]byte)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, []ids.ShortID, ...rpc.Option) error); ok {
		r2 = rf(ctx, addrs, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetTx provides a mock function with given fields: ctx, txID, options
func (_m *PChainClient) GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, txID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, ids.ID, ...rpc.Option) []byte); ok {
		r0 = rf(ctx, txID, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ids.ID, ...rpc.Option) error); ok {
		r1 = rf(ctx, txID, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTxFee provides a mock function with given fields: _a0, _a1
func (_m *PChainClient) GetTxFee(_a0 context.Context, _a1 ...rpc.Option) (*info.GetTxFeeResponse, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *info.GetTxFeeResponse
	if rf, ok := ret.Get(0).(func(context.Context, ...rpc.Option) *info.GetTxFeeResponse); ok {
		r0 = rf(_a0, _a1...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*info.GetTxFeeResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUTXOs provides a mock function with given fields: ctx, addrs, limit, startAddress, startUTXOID, options
func (_m *PChainClient) GetUTXOs(ctx context.Context, addrs []ids.ShortID, limit uint32, startAddress ids.ShortID, startUTXOID ids.ID, options ...rpc.Option) ([][]byte, ids.ShortID, ids.ID, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, addrs, limit, startAddress, startUTXOID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func(context.Context, []ids.ShortID, uint32, ids.ShortID, ids.ID, ...rpc.Option) [][]byte); ok {
		r0 = rf(ctx, addrs, limit, startAddress, startUTXOID, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 ids.ShortID
	if rf, ok := ret.Get(1).(func(context.Context, []ids.ShortID, uint32, ids.ShortID, ids.ID, ...rpc.Option) ids.ShortID); ok {
		r1 = rf(ctx, addrs, limit, startAddress, startUTXOID, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(ids.ShortID)
		}
	}

	var r2 ids.ID
	if rf, ok := ret.Get(2).(func(context.Context, []ids.ShortID, uint32, ids.ShortID, ids.ID, ...rpc.Option) ids.ID); ok {
		r2 = rf(ctx, addrs, limit, startAddress, startUTXOID, options...)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(ids.ID)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(context.Context, []ids.ShortID, uint32, ids.ShortID, ids.ID, ...rpc.Option) error); ok {
		r3 = rf(ctx, addrs, limit, startAddress, startUTXOID, options...)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// IsBootstrapped provides a mock function with given fields: _a0, _a1, _a2
func (_m *PChainClient) IsBootstrapped(_a0 context.Context, _a1 string, _a2 ...rpc.Option) (bool, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, ...rpc.Option) bool); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IssueTx provides a mock function with given fields: ctx, tx, options
func (_m *PChainClient) IssueTx(ctx context.Context, tx []byte, options ...rpc.Option) (ids.ID, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, tx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 ids.ID
	if rf, ok := ret.Get(0).(func(context.Context, []byte, ...rpc.Option) ids.ID); ok {
		r0 = rf(ctx, tx, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ids.ID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []byte, ...rpc.Option) error); ok {
		r1 = rf(ctx, tx, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Peers provides a mock function with given fields: _a0, _a1
func (_m *PChainClient) Peers(_a0 context.Context, _a1 ...rpc.Option) ([]info.Peer, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []info.Peer
	if rf, ok := ret.Get(0).(func(context.Context, ...rpc.Option) []info.Peer); ok {
		r0 = rf(_a0, _a1...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]info.Peer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...rpc.Option) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
