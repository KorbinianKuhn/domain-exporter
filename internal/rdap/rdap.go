package rdap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

var requestTimeout = 30 * time.Second

// Shared clients, so the bootstrap registry is cached between checks
var (
	httpClient      = &http.Client{Timeout: requestTimeout}
	bootstrapClient = &bootstrap.Client{HTTP: httpClient}
	client          = &openrdap.Client{HTTP: httpClient, Bootstrap: bootstrapClient}
)

func isTLDSupported(tld string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	q := &bootstrap.Question{
		RegistryType: bootstrap.DNS,
		Query:        tld,
	}
	answer, err := bootstrapClient.Lookup(q.WithContext(ctx))
	if err != nil {
		return false, err
	}
	return len(answer.URLs) > 0, nil
}

func queryDomainWithConfiguredServer(url string) (*openrdap.Domain, error) {
	// QueryDomain applies its own 30s timeout
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

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &openrdap.Request{
		Type:   openrdap.DomainRequest,
		Query:  domain,
		Server: u,
	}

	resp, err := client.Do(req.WithContext(ctx))
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

func parseResult(domainName string, result *openrdap.Domain) *domain.DomainInfo {
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

	info := domain.NewDomainInfo(domainName)
	info.CheckMethod = domain.CheckMethodRdap
	if len(result.Status) > 0 {
		info.SetStatus(result.Status)
	} else {
		info.SetStatus([]string{"registered"})
	}
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

	return parseResult(url, result), nil
}
