// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ledgerwatch/erigon-lib/state (interfaces: CanonicalsReader)
//
// Generated by this command:
//
//	mockgen -typed=true -destination=./iters_mock.go -package=state . CanonicalsReader
//

// Package state is a generated GoMock package.
package state

import (
	reflect "reflect"

	common "github.com/ledgerwatch/erigon-lib/common"
	kv "github.com/ledgerwatch/erigon-lib/kv"
	order "github.com/ledgerwatch/erigon-lib/kv/order"
	stream "github.com/ledgerwatch/erigon-lib/kv/stream"
	gomock "go.uber.org/mock/gomock"
)

// MockCanonicalsReader is a mock of CanonicalsReader interface.
type MockCanonicalsReader struct {
	ctrl     *gomock.Controller
	recorder *MockCanonicalsReaderMockRecorder
}

// MockCanonicalsReaderMockRecorder is the mock recorder for MockCanonicalsReader.
type MockCanonicalsReaderMockRecorder struct {
	mock *MockCanonicalsReader
}

// NewMockCanonicalsReader creates a new mock instance.
func NewMockCanonicalsReader(ctrl *gomock.Controller) *MockCanonicalsReader {
	mock := &MockCanonicalsReader{ctrl: ctrl}
	mock.recorder = &MockCanonicalsReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCanonicalsReader) EXPECT() *MockCanonicalsReaderMockRecorder {
	return m.recorder
}

// BaseTxnID mocks base method.
func (m *MockCanonicalsReader) BaseTxnID(arg0 kv.Tx, arg1 uint64, arg2 common.Hash) (kv.TxnId, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseTxnID", arg0, arg1, arg2)
	ret0, _ := ret[0].(kv.TxnId)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BaseTxnID indicates an expected call of BaseTxnID.
func (mr *MockCanonicalsReaderMockRecorder) BaseTxnID(arg0, arg1, arg2 any) *MockCanonicalsReaderBaseTxnIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseTxnID", reflect.TypeOf((*MockCanonicalsReader)(nil).BaseTxnID), arg0, arg1, arg2)
	return &MockCanonicalsReaderBaseTxnIDCall{Call: call}
}

// MockCanonicalsReaderBaseTxnIDCall wrap *gomock.Call
type MockCanonicalsReaderBaseTxnIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockCanonicalsReaderBaseTxnIDCall) Return(arg0 kv.TxnId, arg1 error) *MockCanonicalsReaderBaseTxnIDCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockCanonicalsReaderBaseTxnIDCall) Do(f func(kv.Tx, uint64, common.Hash) (kv.TxnId, error)) *MockCanonicalsReaderBaseTxnIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockCanonicalsReaderBaseTxnIDCall) DoAndReturn(f func(kv.Tx, uint64, common.Hash) (kv.TxnId, error)) *MockCanonicalsReaderBaseTxnIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LastFrozenTxNum mocks base method.
func (m *MockCanonicalsReader) LastFrozenTxNum(arg0 kv.Tx) (kv.TxnId, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastFrozenTxNum", arg0)
	ret0, _ := ret[0].(kv.TxnId)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastFrozenTxNum indicates an expected call of LastFrozenTxNum.
func (mr *MockCanonicalsReaderMockRecorder) LastFrozenTxNum(arg0 any) *MockCanonicalsReaderLastFrozenTxNumCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastFrozenTxNum", reflect.TypeOf((*MockCanonicalsReader)(nil).LastFrozenTxNum), arg0)
	return &MockCanonicalsReaderLastFrozenTxNumCall{Call: call}
}

// MockCanonicalsReaderLastFrozenTxNumCall wrap *gomock.Call
type MockCanonicalsReaderLastFrozenTxNumCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockCanonicalsReaderLastFrozenTxNumCall) Return(arg0 kv.TxnId, arg1 error) *MockCanonicalsReaderLastFrozenTxNumCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockCanonicalsReaderLastFrozenTxNumCall) Do(f func(kv.Tx) (kv.TxnId, error)) *MockCanonicalsReaderLastFrozenTxNumCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockCanonicalsReaderLastFrozenTxNumCall) DoAndReturn(f func(kv.Tx) (kv.TxnId, error)) *MockCanonicalsReaderLastFrozenTxNumCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// TxNum2ID mocks base method.
func (m *MockCanonicalsReader) TxNum2ID(arg0 kv.Tx, arg1 uint64, arg2 common.Hash, arg3 uint64) (kv.TxnId, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxNum2ID", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(kv.TxnId)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TxNum2ID indicates an expected call of TxNum2ID.
func (mr *MockCanonicalsReaderMockRecorder) TxNum2ID(arg0, arg1, arg2, arg3 any) *MockCanonicalsReaderTxNum2IDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxNum2ID", reflect.TypeOf((*MockCanonicalsReader)(nil).TxNum2ID), arg0, arg1, arg2, arg3)
	return &MockCanonicalsReaderTxNum2IDCall{Call: call}
}

// MockCanonicalsReaderTxNum2IDCall wrap *gomock.Call
type MockCanonicalsReaderTxNum2IDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockCanonicalsReaderTxNum2IDCall) Return(arg0 kv.TxnId, arg1 error) *MockCanonicalsReaderTxNum2IDCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockCanonicalsReaderTxNum2IDCall) Do(f func(kv.Tx, uint64, common.Hash, uint64) (kv.TxnId, error)) *MockCanonicalsReaderTxNum2IDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockCanonicalsReaderTxNum2IDCall) DoAndReturn(f func(kv.Tx, uint64, common.Hash, uint64) (kv.TxnId, error)) *MockCanonicalsReaderTxNum2IDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// TxnIdsOfCanonicalBlocks mocks base method.
func (m *MockCanonicalsReader) TxnIdsOfCanonicalBlocks(arg0 kv.Tx, arg1, arg2 int, arg3 order.By, arg4 int) (stream.U64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxnIdsOfCanonicalBlocks", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(stream.U64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TxnIdsOfCanonicalBlocks indicates an expected call of TxnIdsOfCanonicalBlocks.
func (mr *MockCanonicalsReaderMockRecorder) TxnIdsOfCanonicalBlocks(arg0, arg1, arg2, arg3, arg4 any) *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxnIdsOfCanonicalBlocks", reflect.TypeOf((*MockCanonicalsReader)(nil).TxnIdsOfCanonicalBlocks), arg0, arg1, arg2, arg3, arg4)
	return &MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall{Call: call}
}

// MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall wrap *gomock.Call
type MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall) Return(arg0 stream.U64, arg1 error) *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall) Do(f func(kv.Tx, int, int, order.By, int) (stream.U64, error)) *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall) DoAndReturn(f func(kv.Tx, int, int, order.By, int) (stream.U64, error)) *MockCanonicalsReaderTxnIdsOfCanonicalBlocksCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
