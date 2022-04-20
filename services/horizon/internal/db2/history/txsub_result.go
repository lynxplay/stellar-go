package history

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/stellar/go/ingest"
	"github.com/stellar/go/xdr"
)

const (
	txSubResultTableName             = "txsub_results"
	txSubResultHashColumnName        = "transaction_hash"
	txSubResultColumnName            = "tx_result"
	txSubResultSubmittedAtColumnName = "submitted_at"
)

// QTxSubmissionResult defines transaction submission result queries.
type QTxSubmissionResult interface {
	GetTxSubmissionResult(ctx context.Context, hash string) (Transaction, error)
	GetTxSubmissionResults(ctx context.Context, hashes []string) ([]Transaction, error)
	SetTxSubmissionResult(ctx context.Context, transaction ingest.LedgerTransaction, sequence uint32, ledgerClosetime time.Time) error
	InitEmptyTxSubmissionResult(ctx context.Context, hash string) error
	DeleteTxSubmissionResultsOlderThan(ctx context.Context, howOldInSeconds uint64) (int64, error)
}

// TxSubGetResult gets the result of a submitted transaction
func (q *Q) GetTxSubmissionResult(ctx context.Context, hash string) (Transaction, error) {
	sql := sq.Select(txSubResultColumnName).
		From(txSubResultTableName).
		Where(sq.NotEq{txSubResultColumnName: nil}).
		Where(sq.Eq{txSubResultHashColumnName: hash})
	var result string
	err := q.Get(ctx, &result, sql)
	if err != nil {
		return Transaction{}, err
	}

	var tx Transaction
	err = json.Unmarshal([]byte(result), &tx)
	return tx, err
}

// TxSubGetResult gets the result of multiple submitted transactions
func (q *Q) GetTxSubmissionResults(ctx context.Context, hashes []string) ([]Transaction, error) {
	sql := sq.Select(txSubResultColumnName).
		From(txSubResultTableName).
		Where(sq.NotEq{txSubResultColumnName: nil}).
		Where(map[string]interface{}{
			txSubResultHashColumnName: hashes,
		})
	var result []string
	err := q.Select(ctx, &result, sql)
	if err != nil {
		return nil, err
	}

	txs := make([]Transaction, len(result))
	for i := 0; i < len(result); i++ {
		err = json.Unmarshal([]byte(result[i]), &txs[i])
	}
	return txs, err
}

// TxSubSetResult sets the result of a submitted transaction
func (q *Q) SetTxSubmissionResult(ctx context.Context, transaction ingest.LedgerTransaction, sequence uint32, ledgerClosetime time.Time) error {
	row, err := transactionToRow(transaction, sequence, xdr.NewEncodingBuffer())
	if err != nil {
		return err
	}
	tx := Transaction{
		LedgerCloseTime:          ledgerClosetime,
		TransactionWithoutLedger: row,
	}
	b, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	sql := sq.Update(txSubResultTableName).
		Where(sq.Eq{txSubResultHashColumnName: row.TransactionHash}).
		SetMap(map[string]interface{}{txSubResultColumnName: b})
	_, err = q.Exec(ctx, sql)
	return err
}

// TxSubInit initializes a submitted transaction
func (q *Q) InitEmptyTxSubmissionResult(ctx context.Context, hash string) error {
	sql := sq.Insert(txSubResultTableName).
		Columns(txSubResultHashColumnName).
		Values(hash).
		Suffix(fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", txSubResultHashColumnName))
	_, err := q.Exec(ctx, sql)
	return err
}

// TxSubDeleteOlderThan deletes entries older than certain duration
func (q *Q) DeleteTxSubmissionResultsOlderThan(ctx context.Context, howOldInSeconds uint64) (int64, error) {
	sql := sq.Delete(txSubResultTableName).
		Where(sq.Expr("now() >= ("+txSubResultSubmittedAtColumnName+" + interval '1 second' * ?)", howOldInSeconds))
	result, err := q.Exec(ctx, sql)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
