package market

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestUniverseSources(t *testing.T) {
	if os.Getenv("LIVE_MARKET") != "1" {
		t.Skip("set LIVE_MARKET=1")
	}
	ctx := context.Background()
	em := NewEastMoney()
	items, total, err := em.List(ctx, 1, 5)
	t.Logf("eastmoney err=%v total=%d n=%d", err, total, len(items))

	time.Sleep(2 * time.Second)
	sl := NewSinaList()
	items2, total2, err2 := sl.List(ctx, 1, 5)
	t.Logf("sina err=%v total=%d n=%d", err2, total2, len(items2))
}
