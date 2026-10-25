package domain

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type DomainStatus int

const (
	StatusUnknown DomainStatus = iota
	StatusFree
	StatusPendingDelete
	StatusRegistered
	StatusActive
)

func (ds DomainStatus) String() string {
	switch ds {
	case StatusUnknown:
		return "unknown"
	case StatusFree:
		return "free"
	case StatusPendingDelete:
		return "pending_delete"
	case StatusRegistered:
		return "registered"
	case StatusActive:
		return "active"
	default:
		return fmt.Sprintf("%d", int(ds))
	}
}

type CheckMethod int

const (
	CheckMethodFailed CheckMethod = iota
	CheckMethodRdap
	CheckMethodWhois
	CheckMethodNSLookup
)

type DomainInfo struct {
	DomainName  string
	status      []string
	ExpiryDate  *time.Time
	CheckMethod CheckMethod
}

func (d CheckMethod) String() string {
	switch d {
	case CheckMethodFailed:
		return "failed"
	case CheckMethodRdap:
		return "rdap"
	case CheckMethodWhois:
		return "whois"
	case CheckMethodNSLookup:
		return "ns_lookup"
	default:
		return fmt.Sprintf("%d", int(d))
	}
}

func NewDomainInfo(domainName string) *DomainInfo {
	return &DomainInfo{
		DomainName:  strings.ToLower(domainName),
		status:      []string{"unknown"},
		ExpiryDate:  nil,
		CheckMethod: CheckMethodFailed,
	}
}

func (d *DomainInfo) SetStatus(status []string) {
	d.status = []string{}
	for _, s := range status {
		d.status = append(d.status, strings.ToLower(s))
	}
}

// EPP (RFC 5731) and RDAP (RFC 8056) status values plus common registry
// specific ones, normalised by normaliseStatus.
var statusMap = map[string]DomainStatus{
	"ok":               StatusActive,
	"active":           StatusActive,
	"connect":          StatusActive, // DENIC
	"inactive":         StatusRegistered,
	"notdelegated":     StatusRegistered,
	"clienthold":       StatusRegistered,
	"serverhold":       StatusRegistered,
	"pendingcreate":    StatusRegistered,
	"pendingrenew":     StatusRegistered,
	"pendingtransfer":  StatusRegistered,
	"pendingupdate":    StatusRegistered,
	"addperiod":        StatusRegistered,
	"autorenewperiod":  StatusRegistered,
	"renewperiod":      StatusRegistered,
	"transferperiod":   StatusRegistered,
	"locked":           StatusRegistered,
	"registered":       StatusRegistered,
	"pendingdelete":    StatusPendingDelete,
	"redemptionperiod": StatusPendingDelete,
	"pendingrestore":   StatusPendingDelete,
	"free":             StatusFree,
	"available":        StatusFree,
}

// Used when combining multiple statuses, the highest wins
var statusPrecedence = map[DomainStatus]int{
	StatusUnknown:       0,
	StatusFree:          1,
	StatusRegistered:    2,
	StatusActive:        3,
	StatusPendingDelete: 4,
}

func normaliseStatus(status string) string {
	s := strings.ToLower(status)
	// Whois often appends a link, e.g. "clientDeleteProhibited https://icann.org/epp#clientDeleteProhibited"
	if i := strings.Index(s, "http"); i >= 0 {
		s = s[:i]
	}
	return strings.NewReplacer(" ", "", "-", "", "_", "").Replace(s)
}

func mapStatus(status string) (DomainStatus, error) {
	s := normaliseStatus(status)

	if mapped, ok := statusMap[s]; ok {
		return mapped, nil
	}

	// clientDeleteProhibited, serverTransferProhibited, ...
	if strings.HasSuffix(s, "prohibited") {
		return StatusRegistered, nil
	}

	// Free text responses of some whois servers
	for _, free := range []string{"notregistered", "nomatch", "notfound", "available", "free"} {
		if strings.Contains(s, free) {
			return StatusFree, nil
		}
	}
	if strings.Contains(s, "registered") {
		return StatusRegistered, nil
	}

	return StatusUnknown, fmt.Errorf("unknown status: %s", status)
}

func (d *DomainInfo) Status() DomainStatus {
	status := StatusUnknown
	isFree := false

	for _, s := range d.status {
		mappedStatus, err := mapStatus(s)
		if err != nil {
			slog.Warn("Unknown status to map", "status", s)
			continue
		}

		if mappedStatus == StatusFree {
			isFree = true
		}

		if statusPrecedence[mappedStatus] > statusPrecedence[status] {
			status = mappedStatus
		}
	}

	if isFree && status != StatusFree {
		slog.Warn("Conflicting status found", "domain", d.DomainName, "status", d.status)
	}

	if status == StatusUnknown && d.ExpiryDate != nil {
		status = StatusRegistered
	}

	return status
}
