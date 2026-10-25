package metrics

import (
	"testing"
	"time"

	"github.com/korbiniankuhn/domain-exporter/internal/domain"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSetDomainStatusExpiry(t *testing.T) {
	m := NewMetrics()
	expiry := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	registered := domain.NewDomainInfo("example.com")
	registered.SetStatus([]string{"active"})
	registered.ExpiryDate = &expiry
	m.SetDomainStatus(0.1, registered)

	if got := testutil.ToFloat64(m.expiresAt.WithLabelValues("example.com")); got != float64(expiry.Unix()) {
		t.Errorf("expires_at = %v, want %v", got, expiry.Unix())
	}

	// Failed check without expiry keeps the last known value
	failed := domain.NewDomainInfo("example.com")
	m.SetDomainStatus(0.1, failed)

	if got := testutil.CollectAndCount(m.expiresAt); got != 1 {
		t.Errorf("expires_at series = %d after failed check, want 1", got)
	}

	// Free domain removes the expiry
	free := domain.NewDomainInfo("example.com")
	free.SetStatus([]string{"free"})
	m.SetDomainStatus(0.1, free)

	if got := testutil.CollectAndCount(m.expiresAt); got != 0 {
		t.Errorf("expires_at series = %d after domain became free, want 0", got)
	}
}
