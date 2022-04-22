package processors

import (
	"context"
	"fmt"
	"time"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/xdr"
)

type processorsRunDurations map[string]time.Duration

func (d processorsRunDurations) AddRunDuration(name string, startTime time.Time) {
	d[name] += time.Since(startTime)
}

type TxSubmissionResultProcessor struct {
	txSubmissionResultQ history.QTxSubmissionResult
	ledger              xdr.LedgerHeaderHistoryEntry
	txs                 []ingest.LedgerTransaction
	processorsRunDurations
}

func (p *TxSubmissionResultProcessor) GetRunDurations() map[string]time.Duration {
	return p.processorsRunDurations
}

func NewTxSubmissionResultProcessor(
	txSubmissionResultQ history.QTxSubmissionResult,
	ledger xdr.LedgerHeaderHistoryEntry,
) *TxSubmissionResultProcessor {
	return &TxSubmissionResultProcessor{
		ledger:                 ledger,
		txSubmissionResultQ:    txSubmissionResultQ,
		processorsRunDurations: make(map[string]time.Duration),
	}
}

func (p *TxSubmissionResultProcessor) ProcessTransaction(ctx context.Context, transaction ingest.LedgerTransaction) (err error) {
	p.txs = append(p.txs, transaction)

	return nil
}

func (p *TxSubmissionResultProcessor) Commit(ctx context.Context) error {
	seq := uint32(p.ledger.Header.LedgerSeq)
	closeTime := time.Unix(int64(p.ledger.Header.ScpValue.CloseTime), 0).UTC()
	startTime := time.Now()
	_, err := p.txSubmissionResultQ.SetTxSubmissionResults(ctx, p.txs, seq, closeTime)
	p.AddRunDuration(fmt.Sprintf("%T", p), startTime)
	return err
}
