package rdap

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/korbiniankuhn/domain-exporter/internal/domain"
	openrdap "github.com/openrdap/rdap"
)

func TestQueryDomainWithCustomServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domain/example.de" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/rdap+json")
		w.Write([]byte(`{"objectClassName": "domain", "ldhName": "example.de", "status": ["active"]}`))
	}))
	defer server.Close()

	d, err := queryDomainWithCustomServer(server.URL, "example.de")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.LDHName != "example.de" {
		t.Errorf("LDHName = %q, want example.de", d.LDHName)
	}
}

func TestQueryDomainWithCustomServerTimeout(t *testing.T) {
	unblock := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-unblock
	}))
	defer server.Close()
	defer close(unblock)

	originalTimeout := requestTimeout
	requestTimeout = 100 * time.Millisecond
	defer func() { requestTimeout = originalTimeout }()

	startedAt := time.Now()
	_, err := queryDomainWithCustomServer(server.URL, "example.de")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(startedAt); elapsed > 2*time.Second {
		t.Errorf("request took %s, expected to time out after %s", elapsed, requestTimeout)
	}
}

func TestParseResult(t *testing.T) {
	result := &openrdap.Domain{
		LDHName: "",
		Status:  []string{"client transfer prohibited", "active"},
		Events: []openrdap.Event{
			{Action: "registration", Date: "2000-01-01T00:00:00Z"},
			{Action: "expiration", Date: "2030-01-01T00:00:00Z"},
		},
	}

	info := parseResult("Example.com", result)

	if info.DomainName != "example.com" {
		t.Errorf("DomainName = %q, want example.com", info.DomainName)
	}
	if info.Status() != domain.StatusActive {
		t.Errorf("Status() = %s, want active", info.Status())
	}
	if info.ExpiryDate == nil || !info.ExpiryDate.Equal(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("ExpiryDate = %v, want 2030-01-01", info.ExpiryDate)
	}
}

func TestParseResultWithoutStatus(t *testing.T) {
	info := parseResult("example.com", &openrdap.Domain{LDHName: "example.com"})

	if info.Status() != domain.StatusRegistered {
		t.Errorf("Status() = %s, want registered", info.Status())
	}
}
