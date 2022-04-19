package history

import (
	"context"
	"time"

	"github.com/stellar/go/ingest"
	"github.com/stretchr/testify/mock"
)

// MockQTxSubmissionResult is a mock implementation of the QTxSubmissionResult interface
type MockQTxSubmissionResult struct {
	mock.Mock
}

func (m *MockQTxSubmissionResult) TxSubGetResult(ctx context.Context, hash string) (Transaction, error) {
	a := m.Called(ctx, hash)
	return a.Get(0).(Transaction), a.Error(1)
}

func (m *MockQTxSubmissionResult) TxSubSetResult(ctx context.Context, transaction ingest.LedgerTransaction, sequence uint32, ledgerClosetime time.Time) error {
	a := m.Called(ctx, transaction, sequence)
	return a.Error(0)
}

func (m *MockQTxSubmissionResult) TxSubInit(ctx context.Context, hash string) error {
	a := m.Called(ctx, hash)
	return a.Error(0)
}

func (m *MockQTxSubmissionResult) TxSubDeleteOlderThan(ctx context.Context, howOldInSeconds uint64) (int64, error) {
	a := m.Called(ctx, howOldInSeconds)
	return a.Get(0).(int64), a.Error(1)
}

func (m *MockQTxSubmissionResult) TxSubGetResults(ctx context.Context, hashes []string) ([]Transaction, error) {
	a := m.Called(ctx, hashes)
	return a.Get(0).([]Transaction), a.Error(1)
}
