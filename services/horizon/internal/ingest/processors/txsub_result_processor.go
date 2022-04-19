package processors

import (
	"context"
	"time"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/xdr"
)

type TxSubmissionResultProcessor struct {
	txSubmissionResultQ history.QTxSubmissionResult
	ledger              xdr.LedgerHeaderHistoryEntry
	txs                 []ingest.LedgerTransaction
}

func NewTxSubmissionResultProcessor(
	txSubmissionResultQ history.QTxSubmissionResult,
	ledger xdr.LedgerHeaderHistoryEntry,
) *TxSubmissionResultProcessor {
	return &TxSubmissionResultProcessor{
		ledger:              ledger,
		txSubmissionResultQ: txSubmissionResultQ,
	}
}

func (p *TxSubmissionResultProcessor) ProcessTransaction(ctx context.Context, transaction ingest.LedgerTransaction) (err error) {
	p.txs = append(p.txs, transaction)

	return nil
}

func (p *TxSubmissionResultProcessor) Commit(ctx context.Context) error {
	seq := uint32(p.ledger.Header.LedgerSeq)
	closeTime := time.Unix(int64(p.ledger.Header.ScpValue.CloseTime), 0).UTC()
	for _, tx := range p.txs {
		// TODO: do all of this at once
		if err := p.txSubmissionResultQ.SetTxSubmissionResult(ctx, tx, seq, closeTime); err != nil {
			return err
		}
	}

	return nil
}
