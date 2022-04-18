package history

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/guregu/null"
)

const (
	txSubTableName                   = "txsub_results"
	txSubHashColumnName              = "transaction_hash"
	txSubResultColumnName            = "tx_result"
	txSubResultSubmittedAtColumnName = "submitted_at"
)

// QTxSubmissionResult defines transaction submission result queries.
type QTxSubmissionResult interface {
	TxSubGetResult(ctx context.Context, hash string) (null.String, error)
	TxSubSetResult(ctx context.Context, hash string, result string) error
	TxSubInit(ctx context.Context, hash string) error
	TxSubDeleteOlderThan(ctx context.Context, howOldInSeconds uint) (int64, error)
}

// TxSubGetResult gets the result of a transaction submitted
func (q *Q) TxSubGetResult(ctx context.Context, hash string) (null.String, error) {
	sql := sq.Select(txSubResultColumnName).
		From(txSubTableName).
		Where(sq.Eq{txSubHashColumnName: hash})
	var result null.String
	err := q.Get(ctx, &result, sql)
	return result, err
}

// TxSubSetResult sets the result of a submitted transaction
func (q *Q) TxSubSetResult(ctx context.Context, hash string, result string) error {
	sql := sq.Update(txSubTableName).
		Where(sq.Eq{txSubHashColumnName: hash}).
		SetMap(map[string]interface{}{txSubResultColumnName: result})
	_, err := q.Exec(ctx, sql)
	return err
}

// TxSubInit initializes a submitted transaction
func (q *Q) TxSubInit(ctx context.Context, hash string) error {
	sql := sq.Insert(txSubTableName).
		Columns(txSubHashColumnName).
		Values(hash)
	_, err := q.Exec(ctx, sql)
	return err
}

// TxSubDeleteOlderThan deletes entries older than certain duration
func (q *Q) TxSubDeleteOlderThan(ctx context.Context, howOldInSeconds uint) (int64, error) {
	sql := sq.Delete(txSubTableName).
		Where(sq.Expr("now() >= ("+txSubResultSubmittedAtColumnName+" + interval '1 second' * ?)", howOldInSeconds))
	result, err := q.Exec(ctx, sql)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
