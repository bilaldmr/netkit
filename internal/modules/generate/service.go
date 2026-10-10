package generate

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

func UUID(version, count int) ([]string, error) {
	if version != 4 && version != 7 {
		return nil, errors.New("version must be 4 or 7")
	}
	if count < 1 || count > 100 {
		return nil, errors.New("count must be between 1 and 100")
	}
	out := make([]string, count)
	for i := range out {
		var b [16]byte
		rand.Read(b[:])
		if version == 7 {
			ms := uint64(time.Now().UnixMilli())
			var ts [8]byte
			binary.BigEndian.PutUint64(ts[:], ms)
			copy(b[:6], ts[2:])
		}
		b[6] = b[6]&0x0f | byte(version)<<4
		b[8] = b[8]&0x3f | 0x80
		h := hex.EncodeToString(b[:])
		out[i] = h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
	}
	return out, nil
}

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	symbolChars = "!@#$%^&*()-_=+[]{};:,.?/"
)

func Password(length, count int, lower, upper, digits, symbols bool) ([]string, error) {
	if length < 8 || length > 128 {
		return nil, errors.New("length must be between 8 and 128")
	}
	if count < 1 || count > 50 {
		return nil, errors.New("count must be between 1 and 50")
	}
	var classes []string
	for _, c := range []struct {
		on    bool
		chars string
	}{{lower, lowerChars}, {upper, upperChars}, {digits, digitChars}, {symbols, symbolChars}} {
		if c.on {
			classes = append(classes, c.chars)
		}
	}
	if len(classes) == 0 {
		return nil, errors.New("select at least one character set")
	}
	all := strings.Join(classes, "")
	out := make([]string, count)
	for i := range out {
		for {
			b := make([]byte, length)
			for j := range b {
				n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(all))))
				b[j] = all[n.Int64()]
			}
			if hasAll(string(b), classes) {
				out[i] = string(b)
				break
			}
		}
	}
	return out, nil
}

func hasAll(s string, classes []string) bool {
	for _, c := range classes {
		if !strings.ContainsAny(s, c) {
			return false
		}
	}
	return true
}

type Time struct {
	Input    string `json:"input"`
	Unix     int64  `json:"unix"`
	UnixMs   int64  `json:"unix_ms"`
	UTC      string `json:"utc"`
	Local    string `json:"local"`
	Timezone string `json:"timezone"`
	RFC1123  string `json:"rfc1123"`
	Relative string `json:"relative"`
}

var layouts = []string{time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", time.RFC1123, time.RFC1123Z}

func Timestamp(value, tz string, now time.Time) (Time, error) {
	loc := time.UTC
	if tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			return Time{}, errors.New("unknown timezone " + strconv.Quote(tz))
		}
		loc = l
	}
	value = strings.TrimSpace(value)
	t, err := parseTime(value, loc, now)
	if err != nil {
		return Time{}, err
	}
	return Time{
		Input:    value,
		Unix:     t.Unix(),
		UnixMs:   t.UnixMilli(),
		UTC:      t.UTC().Format(time.RFC3339),
		Local:    t.In(loc).Format(time.RFC3339),
		Timezone: loc.String(),
		RFC1123:  t.In(loc).Format(time.RFC1123),
		Relative: relative(t, now),
	}, nil
}

func parseTime(v string, loc *time.Location, now time.Time) (time.Time, error) {
	if v == "" || v == "now" {
		return now, nil
	}
	if n, err := strconv.ParseInt(v, 10, 64); err == nil {
		switch d := len(strings.TrimLeft(v, "-")); {
		case d >= 19:
			return time.Unix(0, n), nil
		case d >= 16:
			return time.UnixMicro(n), nil
		case d >= 13:
			return time.UnixMilli(n), nil
		default:
			return time.Unix(n, 0), nil
		}
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, v, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unrecognized time " + strconv.Quote(v) + "; use a unix timestamp or ISO 8601")
}

func relative(t, now time.Time) string {
	d := now.Sub(t)
	suffix := " ago"
	if d < 0 {
		d, suffix = -d, " from now"
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + " minutes" + suffix
	case d < 48*time.Hour:
		return strconv.Itoa(int(d.Hours())) + " hours" + suffix
	case d < 730*24*time.Hour:
		return strconv.Itoa(int(d.Hours()/24)) + " days" + suffix
	default:
		return strconv.Itoa(int(d.Hours()/24/365)) + " years" + suffix
	}
}
