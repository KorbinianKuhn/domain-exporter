package domain

import (
	"testing"
	"time"
)

func TestMapStatus(t *testing.T) {
	tests := []struct {
		status string
		want   DomainStatus
	}{
		// EPP (whois)
		{"ok", StatusActive},
		{"inactive", StatusRegistered},
		{"clientHold", StatusRegistered},
		{"clientDeleteProhibited", StatusRegistered},
		{"serverDeleteProhibited", StatusRegistered},
		{"clientTransferProhibited", StatusRegistered},
		{"clientUpdateProhibited https://icann.org/epp#clientUpdateProhibited", StatusRegistered},
		{"autoRenewPeriod", StatusRegistered},
		{"pendingDelete", StatusPendingDelete},
		{"redemptionPeriod", StatusPendingDelete},
		{"not delegated", StatusRegistered},
		// RDAP
		{"active", StatusActive},
		{"client delete prohibited", StatusRegistered},
		{"server hold", StatusRegistered},
		{"pending delete", StatusPendingDelete},
		{"redemption period", StatusPendingDelete},
		{"pending restore", StatusPendingDelete},
		// Free text
		{"connect", StatusActive},
		{"free", StatusFree},
		{"AVAILABLE", StatusFree},
		{"No match for domain", StatusFree},
		{"not registered", StatusFree},
		{"Registered", StatusRegistered},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got, err := mapStatus(tt.status)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("mapStatus(%q) = %s, want %s", tt.status, got, tt.want)
			}
		})
	}
}

func TestMapStatusUnknown(t *testing.T) {
	if _, err := mapStatus("something else"); err == nil {
		t.Error("expected error for unknown status")
	}
}

func TestStatus(t *testing.T) {
	expiry := time.Now()

	tests := []struct {
		name   string
		status []string
		expiry *time.Time
		want   DomainStatus
	}{
		{"locks only", []string{"clientTransferProhibited", "clientDeleteProhibited"}, nil, StatusRegistered},
		{"locks reversed", []string{"clientDeleteProhibited", "clientTransferProhibited"}, nil, StatusRegistered},
		{"active with locks", []string{"client delete prohibited", "active", "client transfer prohibited"}, nil, StatusActive},
		{"inactive with redemption", []string{"redemption period", "inactive"}, nil, StatusPendingDelete},
		{"pending delete with locks", []string{"pendingDelete", "clientDeleteProhibited"}, nil, StatusPendingDelete},
		{"unknown only", []string{"something else"}, nil, StatusUnknown},
		{"unknown with expiry", []string{"something else"}, &expiry, StatusRegistered},
		{"no status", []string{}, nil, StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDomainInfo("example.com")
			d.SetStatus(tt.status)
			d.ExpiryDate = tt.expiry
			if got := d.Status(); got != tt.want {
				t.Errorf("Status() = %s, want %s", got, tt.want)
			}
		})
	}
}
