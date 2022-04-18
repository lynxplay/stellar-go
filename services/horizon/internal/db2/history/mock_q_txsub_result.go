package history

import (
	"context"

	"github.com/guregu/null"
	"github.com/stretchr/testify/mock"
)

// MockQTxSubmissionResult is a mock implementation of the QTxSubmissionResult interface
type MockQTxSubmissionResult struct {
	mock.Mock
}

func (m MockQTxSubmissionResult) TxSubGetResult(ctx context.Context, hash string) (null.String, error) {
	a := m.Called(ctx, hash)
	return a.Get(0).(null.String), a.Error(1)
}

func (m MockQTxSubmissionResult) TxSubSetResult(ctx context.Context, hash string, result string) error {
	a := m.Called(ctx, hash, result)
	return a.Error(0)
}

func (m MockQTxSubmissionResult) TxSubInit(ctx context.Context, hash string) error {
	a := m.Called(ctx, hash)
	return a.Error(0)
}

func (m MockQTxSubmissionResult) TxSubDeleteOlderThan(ctx context.Context, howOldInSeconds uint) error {
	a := m.Called(ctx, howOldInSeconds)
	return a.Error(0)
}
