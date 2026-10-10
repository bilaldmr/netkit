package cidr

import (
	"errors"
	"math/big"
	"net/netip"
	"strings"
)

type Info struct {
	Input     string `json:"input"`
	CIDR      string `json:"cidr"`
	Version   int    `json:"version"`
	Prefix    int    `json:"prefix"`
	Network   string `json:"network"`
	Broadcast string `json:"broadcast,omitempty"`
	Netmask   string `json:"netmask,omitempty"`
	Wildcard  string `json:"wildcard,omitempty"`
	FirstHost string `json:"first_host"`
	LastHost  string `json:"last_host"`
	Addresses string `json:"addresses"`
	Hosts     string `json:"hosts"`
	IP        string `json:"ip,omitempty"`
	Contains  *bool  `json:"contains,omitempty"`
}

func Calculate(input, ip string) (Info, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Info{}, errors.New("cidr required")
	}
	if !strings.Contains(input, "/") {
		a, err := netip.ParseAddr(input)
		if err != nil {
			return Info{}, err
		}
		input += "/" + map[bool]string{true: "32", false: "128"}[a.Is4()]
	}
	p, err := netip.ParsePrefix(input)
	if err != nil {
		return Info{}, err
	}
	p = p.Masked()
	bits := p.Addr().BitLen()
	hostBits := bits - p.Bits()

	first := p.Addr()
	last := lastAddr(p)
	addresses := new(big.Int).Lsh(big.NewInt(1), uint(hostBits))
	hosts := new(big.Int).Set(addresses)

	info := Info{
		Input:   input,
		CIDR:    p.String(),
		Version: map[bool]int{true: 4, false: 6}[first.Is4()],
		Prefix:  p.Bits(),
		Network: first.String(),
	}
	if first.Is4() {
		mask := netip.AddrFrom4(maskBytes4(p.Bits()))
		wild := mask.As4()
		for i := range wild {
			wild[i] = ^wild[i]
		}
		info.Netmask, info.Wildcard = mask.String(), netip.AddrFrom4(wild).String()
		info.Broadcast = last.String()
		if hostBits >= 2 {
			first, last = first.Next(), last.Prev()
			hosts.Sub(hosts, big.NewInt(2))
		}
	}
	info.FirstHost, info.LastHost = first.String(), last.String()
	info.Addresses, info.Hosts = addresses.String(), hosts.String()

	if ip = strings.TrimSpace(ip); ip != "" {
		a, err := netip.ParseAddr(ip)
		if err != nil {
			return Info{}, err
		}
		c := p.Contains(a.Unmap())
		info.IP, info.Contains = a.String(), &c
	}
	return info, nil
}

func lastAddr(p netip.Prefix) netip.Addr {
	b := p.Addr().AsSlice()
	for i := p.Bits(); i < len(b)*8; i++ {
		b[i/8] |= 1 << (7 - i%8)
	}
	a, _ := netip.AddrFromSlice(b)
	return a
}

func maskBytes4(bits int) [4]byte {
	var m [4]byte
	for i := range bits {
		m[i/8] |= 1 << (7 - i%8)
	}
	return m
}
