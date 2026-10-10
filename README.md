# netkit

Network and developer tools in a single Go binary. Every tool has a web UI and a JSON API.
Standard library only, no external dependencies.

## Tools

| Tool             | Path             | What it does                                                            |
| ---------------- | ---------------- | ----------------------------------------------------------------------- |
| CIDR Calculator  | `/cidr`          | Subnet range, mask, host count, IP-in-range check (IPv4 and IPv6)       |
| DNS Checker      | `/dns`           | A, AAAA, CNAME, MX, NS, TXT records, propagation, SPF/DMARC/DKIM checks |
| Encode / Decode  | `/encode-decode` | Base64, URL encoding, JWT decoding, hashes                              |
| Generators       | `/generate`      | UUIDs, passwords, timestamp conversion, cron next run times             |
| HTTP Headers     | `/headers`       | Response headers by group, security grade and cookie flags              |
| IP Checker       | `/ip`            | IP classification, reverse DNS and location                             |
| JSON Viewer      | `/json`          | Pretty-print JSON from a URL, `?data=` or a POST body                   |
| My IP            | `/myip`          | Your IPv4/IPv6 addresses, user agent and request headers                |
| Port Checker     | `/ports`         | Open TCP ports of a public host (up to 20 per check)                    |
| Redirect Checker | `/redirects`     | Redirect chain hop by hop (max 10)                                      |
| Site Files       | `/site-files`    | robots.txt, sitemap.xml, security.txt and llms.txt                      |
| SSL Checker      | `/ssl`           | Certificate check, supported TLS versions and ciphers, two-host compare |
| WHOIS / RDAP     | `/whois`         | Registration data for a domain, IP or ASN                               |

`GET /api` lists every module with its endpoints and an example request.

Tools that make outbound requests only connect to public IP addresses, also after redirects.

## Running

With Go:

```sh
go run ./cmd/server
```

With Docker:

```sh
cp .env.example .env
make docker-up
```

Then open http://localhost:2347. Run `make` to see all targets.

## Configuration

All settings are environment variables and optional.

| Variable           | Default   | Description                                                      |
| ------------------ | --------- | ---------------------------------------------------------------- |
| `HTTP_SERVER_HOST` | `0.0.0.0` | Listen address                                                   |
| `HTTP_SERVER_PORT` | `2347`    | Listen port                                                      |
| `LOG_LEVEL`        | `info`    | `debug`, `info`, `warn`, `error`                                 |
| `LOG_FORMAT`       | `text`    | `text` or `json`                                                 |
| `IPINFO_TOKEN`     |           | [ipinfo.io](https://ipinfo.io) token for IP Checker locations    |
| `MYIP_V4_URL`      |           | Base URL of this service reachable over IPv4 only, used by My IP |
| `MYIP_V6_URL`      |           | Base URL of this service reachable over IPv6 only, used by My IP |

## Development

```sh
gofmt -l .
go vet ./...
go test ./...
```

Each tool is a module under `internal/modules/<name>` with its routes, service and UI
(`ui/<view>/view.html`, `view.js`, `view.css`), registered in `cmd/server/main.go`.

## License

[MIT](LICENSE)
