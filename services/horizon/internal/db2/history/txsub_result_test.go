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

	resultRet, err := q.TxSubGetResult(ctx, hash)
	tt.Assert.NoError(err)
	tt.Assert.Equal(null.String{}, resultRet)

	result := "yeah!"

	tt.Assert.NoError(q.TxSubSetResult(ctx, hash, result))
	resultRet, err = q.TxSubGetResult(ctx, hash)
	tt.Assert.NoError(err)
	tt.Assert.Equal(null.StringFrom(result), resultRet)

	time.Sleep(2 * time.Second)
	rowsAffected, err := q.TxSubDeleteOlderThan(ctx, 1)
	tt.Assert.NoError(err)
	tt.Assert.Equal(int64(1), rowsAffected)

	_, err = q.TxSubGetResult(ctx, hash)
	tt.Assert.Error(err)
}
