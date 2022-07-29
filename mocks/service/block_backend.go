// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package chain

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/coinbase/rosetta-sdk-go/types"
)

// BlockBackend is an autogenerated mock type for the BlockBackend type
type BlockBackend struct {
	mock.Mock
}

// Block provides a mock function with given fields: ctx, request
func (_m *BlockBackend) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	ret := _m.Called(ctx, request)

	var r0 *types.BlockResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.BlockRequest) *types.BlockResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockResponse)
		}
	}

	var r1 *types.Error
	if rf, ok := ret.Get(1).(func(context.Context, *types.BlockRequest) *types.Error); ok {
		r1 = rf(ctx, request)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*types.Error)
		}
	}

	return r0, r1
}

// BlockTransaction provides a mock function with given fields: ctx, request
func (_m *BlockBackend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	ret := _m.Called(ctx, request)

	var r0 *types.BlockTransactionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.BlockTransactionRequest) *types.BlockTransactionResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockTransactionResponse)
		}
	}

	var r1 *types.Error
	if rf, ok := ret.Get(1).(func(context.Context, *types.BlockTransactionRequest) *types.Error); ok {
		r1 = rf(ctx, request)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*types.Error)
		}
	}

	return r0, r1
}

// ShouldHandleRequest provides a mock function with given fields: req
func (_m *BlockBackend) ShouldHandleRequest(req interface{}) bool {
	ret := _m.Called(req)

	var r0 bool
	if rf, ok := ret.Get(0).(func(interface{}) bool); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
