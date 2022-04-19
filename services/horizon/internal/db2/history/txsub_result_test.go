package history

import (
	"context"
	"testing"
	"time"

	"github.com/guregu/null"
	"github.com/stellar/go/services/horizon/internal/test"
)

func TestTxSubResult(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetHorizonDB(t, tt.HorizonDB)
	q := &Q{tt.HorizonSession()}

	hash := "foobar"
	ctx := context.Background()

	tt.Assert.NoError(q.TxSubInit(ctx, hash))

	transactionPtr, err := q.TxSubGetResult(ctx, hash)
	tt.Assert.NoError(err)
	tt.Assert.Nil(transactionPtr)

	sequence := uint32(123)
	toInsert := buildLedgerTransaction(tt.T, testTransaction{
		index:         1,
		envelopeXDR:   "AAAAACiSTRmpH6bHC6Ekna5e82oiGY5vKDEEUgkq9CB//t+rAAAAyAEXUhsAADDRAAAAAAAAAAAAAAABAAAAAAAAAAsBF1IbAABX4QAAAAAAAAAA",
		resultXDR:     "AAAAAAAAASwAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAFAAAAAAAAAAA=",
		feeChangesXDR: "AAAAAA==",
		metaXDR:       "AAAAAQAAAAAAAAAA",
		hash:          "19aaa18db88605aedec04659fb45e06f240b022eb2d429e05133e4d53cd945ba",
	})
	ledgerCloseTime := time.Now().UTC().Truncate(time.Second)
	expected := Transaction{
		LedgerCloseTime: ledgerCloseTime,
		TransactionWithoutLedger: TransactionWithoutLedger{
			TotalOrderID:     TotalOrderID{528280981504},
			TransactionHash:  "19aaa18db88605aedec04659fb45e06f240b022eb2d429e05133e4d53cd945ba",
			LedgerSequence:   int32(sequence),
			ApplicationOrder: 1,
			Account:          "GAUJETIZVEP2NRYLUESJ3LS66NVCEGMON4UDCBCSBEVPIID773P2W6AY",
			AccountSequence:  "78621794419880145",
			MaxFee:           200,
			FeeCharged:       300,
			OperationCount:   1,
			TxEnvelope:       "AAAAACiSTRmpH6bHC6Ekna5e82oiGY5vKDEEUgkq9CB//t+rAAAAyAEXUhsAADDRAAAAAAAAAAAAAAABAAAAAAAAAAsBF1IbAABX4QAAAAAAAAAA",
			TxResult:         "AAAAAAAAASwAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAFAAAAAAAAAAA=",
			TxFeeMeta:        "AAAAAA==",
			TxMeta:           "AAAAAQAAAAAAAAAA",
			Signatures:       []string{},
			InnerSignatures:  nil,
			MemoType:         "none",
			Memo:             null.NewString("", false),
			Successful:       true,
			TimeBounds:       TimeBounds{Null: true},
		},
	}

	tt.Assert.NoError(q.TxSubSetResult(ctx, hash, toInsert, sequence, ledgerCloseTime))
	transactionPtr, err = q.TxSubGetResult(ctx, hash)
	tt.Assert.NoError(err)
	tt.Assert.NotNil(transactionPtr)
	transaction := *transactionPtr

	// ignore created time and updated time
	transaction.CreatedAt = expected.CreatedAt
	transaction.UpdatedAt = expected.UpdatedAt

	// compare ClosedAt separately because reflect.DeepEqual does not handle time.Time
	closedAt := transaction.LedgerCloseTime
	transaction.LedgerCloseTime = expected.LedgerCloseTime

	tt.Assert.True(closedAt.Equal(expected.LedgerCloseTime))
	tt.Assert.Equal(transaction, expected)

	time.Sleep(2 * time.Second)
	rowsAffected, err := q.TxSubDeleteOlderThan(ctx, 1)
	tt.Assert.NoError(err)
	tt.Assert.Equal(int64(1), rowsAffected)

	_, err = q.TxSubGetResult(ctx, hash)
	tt.Assert.Error(err)
}
