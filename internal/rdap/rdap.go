package rdap

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/korbiniankuhn/domain-exporter/internal/domain"
	"github.com/korbiniankuhn/domain-exporter/internal/utils"
	openrdap "github.com/openrdap/rdap"
	"github.com/openrdap/rdap/bootstrap"
)

var (
	ErrRDAPNotSupported = fmt.Errorf("RDAP not supported for this TLD")
)

func isTLDSupported(tld string) (bool, error) {
	b := &bootstrap.Client{}
	q := &bootstrap.Question{
		RegistryType: bootstrap.DNS,
		Query:        tld,
	}
	answer, err := b.Lookup(q)
	if err != nil {
		return false, err
	}
	return len(answer.URLs) > 0, nil
}

func queryDomainWithConfiguredServer(url string) (*openrdap.Domain, error) {
	client := &openrdap.Client{}
	domainInfo, err := client.QueryDomain(url)
	if err != nil {
		return nil, err
	}
	return domainInfo, nil
}

func queryDomainWithCustomServer(serverURL, domain string) (*openrdap.Domain, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	req := &openrdap.Request{
		Type:   openrdap.DomainRequest,
		Query:  domain,
		Server: u,
	}

	client := &openrdap.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	d, ok := resp.Object.(*openrdap.Domain)
	if !ok {
		return nil, fmt.Errorf("expected domain object, got %T", resp.Object)
	}

	return d, nil
}

func queryDomain(url string) (*openrdap.Domain, error) {
	tld := utils.GetTLD(url)

	if tld == "de" {
		return queryDomainWithCustomServer("https://rdap.denic.de", url)
	}

	isSupported, err := isTLDSupported(tld)
	if err != nil {
		return nil, err
	}

	if isSupported {
		return queryDomainWithConfiguredServer(url)
	}

	return nil, ErrRDAPNotSupported
}

func parseResult(result *openrdap.Domain) *domain.DomainInfo {
	status := strings.Join(result.Status, ", ")
	var expiryDate *time.Time

	for _, ev := range result.Events {
		if strings.Contains(strings.ToLower(ev.Action), "expiration") {
			date, err := time.Parse(time.RFC3339, ev.Date)
			if err == nil {
				expiryDate = &date
			} else {
				slog.Warn("Could not parse expiry date", "date", ev.Date, "error", err)
			}
		}
	}

	info := domain.NewDomainInfo(result.LDHName)
	info.CheckMethod = domain.CheckMethodRdap
	info.SetStatus([]string{status})
	info.ExpiryDate = expiryDate

	return info
}

func CheckDomain(url string) (*domain.DomainInfo, error) {
	result, err := queryDomain(url)

	info := domain.NewDomainInfo(url)
	info.CheckMethod = domain.CheckMethodRdap

	if cerr, ok := err.(*openrdap.ClientError); ok {
		if cerr.Type == openrdap.ObjectDoesNotExist {
			info.SetStatus([]string{"free"})
			return info, nil
		}
	}

	if err != nil {
		return nil, err
	}

	return parseResult(result), nil
}
