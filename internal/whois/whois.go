package whois

import (
	"strings"

	"github.com/korbiniankuhn/domain-exporter/internal/domain"
	"github.com/korbiniankuhn/domain-exporter/internal/utils"
	whoislib "github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

func CheckDomain(url string) (*domain.DomainInfo, error) {
	whois_result, err := whoislib.Whois(url)
	if err != nil {
		return nil, err
	}

	tld := utils.GetTLD(url)

	info := domain.NewDomainInfo(url)
	info.CheckMethod = domain.CheckMethodWhois

	if tld == "at" {
		if strings.Contains(whois_result, "% nothing found") {
			info.SetStatus([]string{"free"})
			return info, nil
		}
	}

	parsedInfo, err := whoisparser.Parse(whois_result)

	if err != nil {
		if err == whoisparser.ErrNotFoundDomain {
			info.SetStatus([]string{"free"})
			return info, nil
		}
		return nil, err
	} else if len(parsedInfo.Domain.Status) > 0 {
		info.SetStatus(parsedInfo.Domain.Status)
	} else if parsedInfo.Registrant != nil || parsedInfo.Registrar != nil {
		info.SetStatus([]string{"registered"})
	}

	info.ExpiryDate = parsedInfo.Domain.ExpirationDateInTime

	return info, nil
}
