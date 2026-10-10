package ports

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/enteresanlikk/netkit/internal/platform"
)

const MaxPorts = 20

var DefaultPorts = []int{21, 22, 25, 53, 80, 110, 143, 443, 465, 587, 993, 995, 3306, 5432, 6379, 8080, 8443}

var serviceNames = map[int]string{
	21: "FTP", 22: "SSH", 23: "Telnet", 25: "SMTP", 53: "DNS", 80: "HTTP", 110: "POP3", 143: "IMAP",
	443: "HTTPS", 465: "SMTPS", 587: "Submission", 993: "IMAPS", 995: "POP3S", 1433: "MSSQL",
	3306: "MySQL", 3389: "RDP", 5432: "PostgreSQL", 5672: "AMQP", 6379: "Redis", 8080: "HTTP-alt",
	8443: "HTTPS-alt", 9200: "Elasticsearch", 27017: "MongoDB",
}

type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Result struct {
	Port      int     `json:"port"`
	Service   string  `json:"service,omitempty"`
	State     string  `json:"state"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
}

type Report struct {
	Host    string   `json:"host"`
	Address string   `json:"address"`
	Results []Result `json:"results"`
}

type Service struct {
	Resolver Resolver
	Dialer   Dialer
}

func (s Service) Check(ctx context.Context, host, portList string) (Report, error) {
	host = cleanHost(host)
	if host == "" {
		return Report{}, errors.New("host required")
	}
	ports, err := ParsePorts(portList)
	if err != nil {
		return Report{}, err
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		addrs, err := s.Resolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return Report{}, err
		}
		if len(addrs) == 0 {
			return Report{}, errors.New("no addresses for " + host)
		}
		addr = addrs[0]
	}
	addr = addr.Unmap()
	if err := platform.CheckPublic(addr.String()); err != nil {
		return Report{}, err
	}

	results := make([]Result, len(ports))
	var wg sync.WaitGroup
	for i, p := range ports {
		wg.Go(func() {
			results[i] = s.probe(ctx, addr, p)
		})
	}
	wg.Wait()
	return Report{Host: host, Address: addr.String(), Results: results}, nil
}

func (s Service) probe(ctx context.Context, addr netip.Addr, port int) Result {
	r := Result{Port: port, Service: serviceNames[port]}
	c, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	start := time.Now()
	conn, err := s.Dialer.DialContext(c, "tcp", netip.AddrPortFrom(addr, uint16(port)).String())
	switch {
	case err == nil:
		conn.Close()
		r.State = "open"
		r.LatencyMs = float64(time.Since(start).Microseconds()) / 1000
	case errors.Is(err, syscall.ECONNREFUSED):
		r.State = "closed"
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded):
		r.State = "filtered"
	default:
		r.State, r.Error = "error", err.Error()
	}
	return r
}

func ParsePorts(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return DefaultPorts, nil
	}
	var out []int
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lo, hi, isRange := strings.Cut(part, "-")
		a, err := parsePort(lo)
		if err != nil {
			return nil, err
		}
		b := a
		if isRange {
			if b, err = parsePort(hi); err != nil {
				return nil, err
			}
			if b < a {
				return nil, errors.New("invalid range " + strconv.Quote(part))
			}
		}
		for p := a; p <= b; p++ {
			if !slices.Contains(out, p) {
				out = append(out, p)
			}
			if len(out) > MaxPorts {
				return nil, errors.New("at most " + strconv.Itoa(MaxPorts) + " ports per check")
			}
		}
	}
	if len(out) == 0 {
		return DefaultPorts, nil
	}
	return out, nil
}

func parsePort(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 || n > 65535 {
		return 0, errors.New("invalid port " + strconv.Quote(s))
	}
	return n, nil
}

func cleanHost(h string) string {
	h = strings.TrimSpace(h)
	if i := strings.Index(h, "://"); i >= 0 {
		h = h[i+3:]
	}
	h, _, _ = strings.Cut(h, "/")
	if host, _, err := net.SplitHostPort(h); err == nil {
		return host
	}
	return strings.Trim(h, "[]")
}
