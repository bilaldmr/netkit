package platform

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"syscall"
	"time"
)

type Deps struct {
	Resolver *net.Resolver
	Dialer   *net.Dialer
	HTTP     *http.Client
	NoFollow *http.Client
}

func NewDeps() Deps {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, _ := net.SplitHostPort(address)
			return CheckPublic(host)
		},
	}
	transport := browserHeaders{&http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		ForceAttemptHTTP2:   true,
	}}
	return Deps{
		Resolver: net.DefaultResolver,
		Dialer:   dialer,
		HTTP:     &http.Client{Timeout: 10 * time.Second, Transport: transport},
		NoFollow: &http.Client{
			Timeout:       10 * time.Second,
			Transport:     transport,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func CheckPublic(host string) error {
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return errors.New("invalid ip " + strconv.Quote(host))
	}
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return errors.New("address " + ip.String() + " is not public")
	}
	return nil
}

var BrowserHeaders = map[string]string{
	"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36",
	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	"Accept-Language":           "en-US,en;q=0.9",
	"Sec-Ch-Ua":                 `"Chromium";v="141", "Not?A_Brand";v="8", "Google Chrome";v="141"`,
	"Sec-Ch-Ua-Mobile":          "?0",
	"Sec-Ch-Ua-Platform":        `"macOS"`,
	"Sec-Fetch-Dest":            "document",
	"Sec-Fetch-Mode":            "navigate",
	"Sec-Fetch-Site":            "none",
	"Sec-Fetch-User":            "?1",
	"Upgrade-Insecure-Requests": "1",
}

type browserHeaders struct{ http.RoundTripper }

func (b browserHeaders) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	for k, v := range BrowserHeaders {
		if r.Header.Get(k) == "" {
			r.Header.Set(k, v)
		}
	}
	return b.RoundTripper.RoundTrip(r)
}
