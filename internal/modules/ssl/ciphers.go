package ssl

import (
	"context"
	"crypto/tls"
	"slices"
	"sync"
	"time"
)

type VersionSupport struct {
	Version   string   `json:"version"`
	Supported bool     `json:"supported"`
	Ciphers   []string `json:"ciphers,omitempty"`
}

type CipherScan struct {
	Host     string           `json:"host"`
	Port     string           `json:"port"`
	Error    string           `json:"error,omitempty"`
	Versions []VersionSupport `json:"versions"`
}

type CipherReport struct {
	Hosts          []CipherScan `json:"hosts"`
	CommonVersions []string     `json:"common_versions,omitempty"`
	CommonCiphers  []string     `json:"common_ciphers,omitempty"`
	Weak           []string     `json:"weak"`
}

var scanVersions = []uint16{tls.VersionTLS13, tls.VersionTLS12, tls.VersionTLS11, tls.VersionTLS10}

var scanSuites = append(tls.CipherSuites(), tls.InsecureCipherSuites()...)

var allSuiteIDs = func() []uint16 {
	var ids []uint16
	for _, cs := range scanSuites {
		ids = append(ids, cs.ID)
	}
	return ids
}()

const probeTimeout = 4 * time.Second

func (s Service) Ciphers(ctx context.Context, target, compare string) (CipherReport, error) {
	targets := []string{target}
	if compare != "" {
		targets = append(targets, compare)
	}
	type hp struct{ host, port string }
	var hps []hp
	for _, t := range targets {
		host, port, err := splitTarget(t)
		if err != nil {
			return CipherReport{}, err
		}
		hps = append(hps, hp{host, port})
	}

	rep := CipherReport{Hosts: make([]CipherScan, len(hps))}
	for _, cs := range tls.InsecureCipherSuites() {
		rep.Weak = append(rep.Weak, cs.Name)
	}
	var wg sync.WaitGroup
	for i, h := range hps {
		wg.Go(func() { rep.Hosts[i] = s.scan(ctx, h.host, h.port) })
	}
	wg.Wait()

	if len(rep.Hosts) == 2 {
		a, b := rep.Hosts[0], rep.Hosts[1]
		for i, va := range a.Versions {
			vb := b.Versions[i]
			if va.Supported && vb.Supported {
				rep.CommonVersions = append(rep.CommonVersions, va.Version)
			}
			for _, c := range va.Ciphers {
				if slices.Contains(vb.Ciphers, c) && !slices.Contains(rep.CommonCiphers, c) {
					rep.CommonCiphers = append(rep.CommonCiphers, c)
				}
			}
		}
	}
	return rep, nil
}

func (s Service) scan(ctx context.Context, host, port string) CipherScan {
	sc := CipherScan{Host: host, Port: port, Versions: make([]VersionSupport, len(scanVersions))}
	probe := func(cfg *tls.Config) (tls.ConnectionState, error) {
		ctx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		return s.Prober.Probe(ctx, host, port, cfg)
	}

	var wg sync.WaitGroup
	errs := make([]error, len(scanVersions))
	for i, v := range scanVersions {
		wg.Go(func() {
			st, err := probe(&tls.Config{MinVersion: v, MaxVersion: v, CipherSuites: allSuiteIDs})
			sc.Versions[i] = VersionSupport{Version: tls.VersionName(v), Supported: err == nil}
			errs[i] = err
			if err == nil && v == tls.VersionTLS13 {
				sc.Versions[i].Ciphers = []string{tls.CipherSuiteName(st.CipherSuite)}
			}
		})
	}
	wg.Wait()
	if !slices.ContainsFunc(sc.Versions, func(v VersionSupport) bool { return v.Supported }) {
		sc.Error = errs[1].Error()
		return sc
	}

	var mu sync.Mutex
	sem := make(chan struct{}, 8)
	for i, v := range scanVersions {
		if v == tls.VersionTLS13 || !sc.Versions[i].Supported {
			continue
		}
		for _, cs := range scanSuites {
			if !slices.Contains(cs.SupportedVersions, v) {
				continue
			}
			wg.Go(func() {
				sem <- struct{}{}
				defer func() { <-sem }()
				if _, err := probe(&tls.Config{MinVersion: v, MaxVersion: v, CipherSuites: []uint16{cs.ID}}); err == nil {
					mu.Lock()
					sc.Versions[i].Ciphers = append(sc.Versions[i].Ciphers, cs.Name)
					mu.Unlock()
				}
			})
		}
	}
	wg.Wait()
	for i := range sc.Versions {
		slices.Sort(sc.Versions[i].Ciphers)
	}
	return sc
}
