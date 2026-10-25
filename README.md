# Domain Exporter

Prometheus domain details exporter using RDAP, Whois and NS Lookup.

## Getting Started

1. Create a `config.yml` and `docker-compose.yml`: Examples below
2. Run `docker compose up`

## Limitations

- ✅ Domain status is requested through RDAP, falls back to Whois and as NS lookup as last resort
- 📅 Expiry Date is only available on a subset of top-level domains due to privacy constraints
- 🔍 The NS lookup fallback can only tell whether a domain is delegated: a registered domain without delegation (e.g. `clientHold`) is reported as free


## Configuration

Configuration can be passed through a config.yml or as environment variables (`CHECK_INTERVAL_IN_SECONDS`, `LOGGING_LEVEL`, `LOGGING_FORMAT`). Environment variables take precedence. Domains can only be set in the config.yml.

config.yml

```yaml
check_interval_in_seconds: 86400

logging:
  level: info # one of (debug, info, warn, error)
  format: text # one of (text, json, console)

domains:
  - example.com
```

docker-compose.yml

```yaml
services:
  domain-exporter:
    image: ghcr.io/korbiniankuhn/domain-exporter:1.0.0
    container_name: domain-exporter
    ports:
      - 127.0.0.1:2112:2112
    volumes:
      - ./config.yml:/domain-exporter/config.yml:ro
    read_only: true
    cap_drop:
      - ALL
    security_opt:
      - no-new-privileges:true
```

The image is based on [distroless](https://github.com/GoogleContainerTools/distroless) and runs as non-root user `65532`, so the mounted `config.yml` must be readable by that user (e.g. `chmod 644 config.yml`).

## Monitoring

Prometheus metrics are exported under [localhost:2112/metrics](localhost:2112/metrics)

domain_status: 0=Unknown, 1=Free, 2=PendingDelete, 3=Registered, 4=Active

domain_probe_method: 0=Failed, 1=RDAP, 2=Whois, 3=NS_Lookup

```yaml
# HELP domain_check_timestamp_seconds Unix timestamp of the last check
# TYPE domain_check_timestamp_seconds gauge
domain_check_timestamp_seconds 1.759771491978996e+09
# HELP domain_check_total Total number of domain checks
# TYPE domain_check_total counter
domain_check_total 4
# HELP domain_expires_at Expiry date of the domain as a Unix timestamp
# TYPE domain_expires_at gauge
domain_expires_at{domain="example.com"} 1.7865936e+09
# HELP domain_probe_duration_seconds Duration in seconds of the domain probe
# TYPE domain_probe_duration_seconds gauge
domain_probe_duration_seconds{domain="example-maybe-free.eu"} 1.43701925
domain_probe_duration_seconds{domain="example.ch"} 0.375900791
domain_probe_duration_seconds{domain="example.com"} 0.153589416
domain_probe_duration_seconds{domain="example.de"} 0.107726875
# HELP domain_probe_method The method used for probing the domain
# TYPE domain_probe_method gauge
domain_probe_method{domain="example-maybe-free.eu"} 2
domain_probe_method{domain="example.ch"} 3
domain_probe_method{domain="example.com"} 1
domain_probe_method{domain="example.de"} 1
# HELP domain_status Status of the domain
# TYPE domain_status gauge
domain_status{domain="example-maybe-free.eu"} 1
domain_status{domain="example.ch"} 4
domain_status{domain="example.com"} 3
domain_status{domain="example.de"} 4
```

## Grafana

A prebuilt dashboard is [here](dashboard.json)

![Screenshot of the Grafana dashboard for Domain Exporter](dashboard.png)

## SBOM & Attestations

Every image is published with a signed SLSA build provenance and SPDX SBOM attestation. You can verify them with the [GitHub CLI](https://cli.github.com/):

```sh
# Build provenance
gh attestation verify oci://ghcr.io/korbiniankuhn/domain-exporter:latest --owner KorbinianKuhn

# SBOM
gh attestation verify oci://ghcr.io/korbiniankuhn/domain-exporter:latest --owner KorbinianKuhn \
  --predicate-type https://spdx.dev/Document/v2.3
```

The image also contains per-platform BuildKit SBOM and provenance attestations:

```sh
docker buildx imagetools inspect ghcr.io/korbiniankuhn/domain-exporter:latest --format '{{ json .SBOM }}'
docker buildx imagetools inspect ghcr.io/korbiniankuhn/domain-exporter:latest --format '{{ json .Provenance }}'
```

The SBOM is also attached as `sbom.spdx.json` to every workflow run.

## Credits

- [https://github.com/likexian/whois](https://github.com/likexian/whois)
- [https://github.com/likexian/whois-parser](https://github.com/likexian/whois-parser)
- [https://github.com/openrdap/rdap](https://github.com/openrdap/rdap)
- [https://github.com/spf13/viper](https://github.com/spf13/viper)
