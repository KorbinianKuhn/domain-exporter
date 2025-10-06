package ns_lookup

import (
	"net"

	"github.com/korbiniankuhn/domain-exporter/internal/domain"
)

func CheckDomain(url string) (*domain.DomainInfo, error) {
	info := domain.NewDomainInfo(url)
	info.CheckMethod = domain.CheckMethodNSLookup

	ns, err := net.LookupNS(url)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			info.SetStatus([]string{"free"})
			return info, nil
		}
		return nil, err
	}

	if len(ns) == 0 {
		info.SetStatus([]string{"registered"})
	} else {
		info.SetStatus([]string{"active"})
	}

	return info, nil
}
