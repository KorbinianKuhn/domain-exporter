package domain

import (
	"fmt"
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

func (d *DomainInfo) Status() DomainStatus {
	if d.ExpiryDate != nil {
		return StatusRegistered
	}

	for _, s := range d.status {
		if strings.Contains(s, "free") {
			return StatusFree
		}
		if strings.Contains(s, "active") {
			return StatusActive
		}
		if strings.Contains(s, "registered") {
			return StatusRegistered
		}
		if strings.Contains(s, "pending delete") {
			return StatusPendingDelete
		}
	}

	return StatusUnknown
}
