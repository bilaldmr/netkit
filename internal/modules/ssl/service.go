package ssl

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strings"
	"time"
)

type Prober interface {
	Probe(ctx context.Context, host, port string, cfg *tls.Config) (tls.ConnectionState, error)
}

type Cert struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	DNSNames  []string  `json:"dns_names"`
}

type Report struct {
	Host        string `json:"host"`
	Valid       bool   `json:"valid"`
	VerifyError string `json:"verify_error,omitempty"`
	DaysLeft    int    `json:"days_left"`
	TLSVersion  string `json:"tls_version"`
	Cipher      string `json:"cipher"`
	Chain       []Cert `json:"chain"`
}

type Service struct{ Prober Prober }

func (s Service) Check(ctx context.Context, target string) (Report, error) {
	host, port, err := splitTarget(target)
	if err != nil {
		return Report{}, err
	}
	st, err := s.Prober.Probe(ctx, host, port, &tls.Config{})
	if err != nil {
		return Report{}, err
	}
	certs := st.PeerCertificates
	if len(certs) == 0 {
		return Report{}, errors.New("no certificate presented")
	}
	inter := x509.NewCertPool()
	for _, c := range certs[1:] {
		inter.AddCert(c)
	}
	rep := Report{
		Host:       host,
		DaysLeft:   int(time.Until(certs[0].NotAfter).Hours() / 24),
		TLSVersion: tls.VersionName(st.Version),
		Cipher:     tls.CipherSuiteName(st.CipherSuite),
	}
	if _, err := certs[0].Verify(x509.VerifyOptions{DNSName: host, Intermediates: inter}); err != nil {
		rep.VerifyError = err.Error()
	}
	rep.Valid = rep.VerifyError == ""
	for _, c := range certs {
		rep.Chain = append(rep.Chain, Cert{c.Subject.String(), c.Issuer.String(), c.NotBefore, c.NotAfter, c.DNSNames})
	}
	return rep, nil
}

func splitTarget(target string) (host, port string, err error) {
	target = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(target), "https://"), "http://")
	target, _, _ = strings.Cut(target, "/")
	if target == "" {
		return "", "", errors.New("host required")
	}
	if host, port, err := net.SplitHostPort(target); err == nil {
		return host, port, nil
	}
	return target, "443", nil
}
