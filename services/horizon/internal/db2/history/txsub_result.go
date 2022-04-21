package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/stellar/go/ingest"
	"github.com/stellar/go/xdr"
)

const (
	txSubResultTableName             = "txsub_results"
	txSubResultHashColumnName        = "transaction_hash"
	txSubResultInnerHashColumnName   = "inner_transaction_hash"
	txSubResultColumnName            = "tx_result"
	txSubResultSubmittedAtColumnName = "submitted_at"
)

// QTxSubmissionResult defines transaction submission result queries.
type QTxSubmissionResult interface {
	GetTxSubmissionResult(ctx context.Context, hash string) (Transaction, error)
	GetTxSubmissionResults(ctx context.Context, hashes []string) ([]Transaction, error)
	SetTxSubmissionResults(ctx context.Context, transactions []ingest.LedgerTransaction, sequence uint32, ledgerClosetime time.Time) (int64, error)
	InitEmptyTxSubmissionResult(ctx context.Context, hash string, innerHash string) error
	DeleteTxSubmissionResultsOlderThan(ctx context.Context, howOldInSeconds uint64) (int64, error)
}

// TxSubGetResult gets the result of a submitted transaction
func (q *Q) GetTxSubmissionResult(ctx context.Context, hash string) (Transaction, error) {
	transactions, err := q.GetTxSubmissionResults(ctx, []string{hash})
	if err != nil {
		return Transaction{}, err
	}
	switch len(transactions) {
	case 0:
		return Transaction{}, sql.ErrNoRows
	case 1:
		return transactions[0], nil
	default:
		return Transaction{}, fmt.Errorf("unexpected result size > 1 (%d)", len(transactions))
	}
}

// TxSubGetResult gets the result of multiple submitted transactions
func (q *Q) GetTxSubmissionResults(ctx context.Context, hashes []string) ([]Transaction, error) {
	byHash := sq.Select(txSubResultColumnName).
		From(txSubResultTableName).
		Where(sq.NotEq{txSubResultColumnName: nil}).
		Where(map[string]interface{}{
			txSubResultHashColumnName: hashes,
		})
	byInnerHash := sq.Select(txSubResultColumnName).
		From(txSubResultTableName).
		Where(sq.NotEq{txSubResultColumnName: nil}).
		Where(map[string]interface{}{
			txSubResultInnerHashColumnName: hashes,
		})
	byInnerHashString, args, err := byInnerHash.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "could not get string for inner hash sql query")
	}
	union := byHash.Suffix("UNION ALL "+byInnerHashString, args...)
	var result []string
	err = q.Select(ctx, &result, union)
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
func (q *Q) SetTxSubmissionResults(ctx context.Context, transactions []ingest.LedgerTransaction, sequence uint32, ledgerClosetime time.Time) (int64, error) {

	if len(transactions) == 0 {
		return 0, nil
	}
	caseStmt := sq.Case(txSubResultHashColumnName)
	hashes := make([]string, len(transactions))
	for i, transaction := range transactions {
		row, err := transactionToRow(transaction, sequence, xdr.NewEncodingBuffer())
		if err != nil {
			return 0, err
		}
		tx := Transaction{
			LedgerCloseTime:          ledgerClosetime,
			TransactionWithoutLedger: row,
		}
		serialized, err := json.Marshal(tx)
		if err != nil {
			return 0, err
		}
		caseStmt = caseStmt.When(sq.Expr("?", row.TransactionHash), sq.Expr("?", serialized))
		hashes[i] = row.TransactionHash
	}

	sql := sq.Update(txSubResultTableName).
		Set(txSubResultColumnName, caseStmt).
		Where(sq.Eq{txSubResultHashColumnName: hashes})
	result, err := q.Exec(ctx, sql)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// TxSubInit initializes a submitted transaction, idempotent, doesn't matter if row with hash already exists.
func (q *Q) InitEmptyTxSubmissionResult(ctx context.Context, hash string, innerHash string) error {
	setMap := map[string]interface{}{
		txSubResultHashColumnName: hash,
	}
	if innerHash != "" {
		setMap[txSubResultInnerHashColumnName] = innerHash
	}
	sql := sq.Insert(txSubResultTableName).
		SetMap(setMap).
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
