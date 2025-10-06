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
	StatusActive
	StatusRegistered
	StatusPendingDelete
	StatusFree
)

func (ds DomainStatus) String() string {
	switch ds {
	case StatusUnknown:
		return "unknown"
	case StatusActive:
		return "active"
	case StatusRegistered:
		return "registered"
	case StatusPendingDelete:
		return "pending_delete"
	case StatusFree:
		return "free"
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

func mapStatus(status string) (DomainStatus, error) {
	s := strings.ReplaceAll(strings.ToLower(status), " ", "")
	if strings.Contains(s, "free") || strings.Contains(s, "available") || strings.Contains(s, "notregistered") || strings.Contains(s, "nomatch") {
		return StatusFree, nil
	}
	if strings.Contains(s, "active") || strings.Contains(s, "ok") {
		return StatusActive, nil
	}
	if strings.Contains(s, "registered") || strings.Contains(s, "inactive") || strings.Contains(s, "create") || strings.Contains(s, "hold") || strings.Contains(s, "clienttransferprohibited") {
		return StatusRegistered, nil
	}
	if strings.Contains(s, "delete") || strings.Contains(s, "redemption") {
		return StatusPendingDelete, nil
	}
	return StatusUnknown, fmt.Errorf("unknown status: %s", status)
}

func (d *DomainInfo) Status() DomainStatus {
	status := StatusUnknown

	for _, s := range d.status {
		mappedStatus, err := mapStatus(s)
		if err != nil {
			slog.Warn("Unknown status to map", "status", s)
			continue
		}

		if status != StatusUnknown && status != mappedStatus {
			slog.Warn("Conflicting status found", "previous", status, "current", mappedStatus)
		}

		status = mappedStatus
	}

	if status == StatusUnknown && d.ExpiryDate != nil {
		status = StatusRegistered
	}

	return status
}
