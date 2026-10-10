package generate

import (
	"errors"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

type cronField struct {
	name     string
	min, max int
	names    []string // names[i] is the alias of value min+i
}

var cronFields = []cronField{
	{"minute", 0, 59, nil},
	{"hour", 0, 23, nil},
	{"day of month", 1, 31, nil},
	{"month", 1, 12, []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}},
	{"day of week", 0, 7, []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}},
}

var cronMacros = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

type cronSpec struct {
	sets             [5]uint64 // bit v set = value v allowed
	domStar, dowStar bool
}

type CronField struct {
	Name   string `json:"name"`
	Values string `json:"values"`
}

type Cron struct {
	Expr     string      `json:"expr"`
	Timezone string      `json:"timezone"`
	Fields   []CronField `json:"fields"`
	Next     []string    `json:"next"`
}

func CronNext(expr, tz string, count int, now time.Time) (Cron, error) {
	if count < 1 || count > 50 {
		return Cron{}, errors.New("count must be between 1 and 50")
	}
	loc := time.UTC
	if tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			return Cron{}, errors.New("unknown timezone " + strconv.Quote(tz))
		}
		loc = l
	}
	expr = strings.TrimSpace(expr)
	spec, err := parseCron(expr)
	if err != nil {
		return Cron{}, err
	}
	c := Cron{Expr: expr, Timezone: loc.String(), Next: []string{}}
	for i, f := range cronFields {
		c.Fields = append(c.Fields, CronField{f.name, describe(spec.sets[i], f)})
	}
	t := now.In(loc)
	for range count {
		if t, err = spec.next(t); err != nil {
			return c, err
		}
		c.Next = append(c.Next, t.Format(time.RFC3339))
	}
	return c, nil
}

func parseCron(expr string) (cronSpec, error) {
	if m, ok := cronMacros[strings.ToLower(expr)]; ok {
		expr = m
	}
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return cronSpec{}, errors.New("expected 5 fields: minute hour day-of-month month day-of-week")
	}
	var s cronSpec
	for i, f := range cronFields {
		set, err := parseField(parts[i], f)
		if err != nil {
			return cronSpec{}, errors.New(f.name + ": " + err.Error())
		}
		s.sets[i] = set
	}
	if s.sets[4]&(1<<7) != 0 { // 7 is Sunday too
		s.sets[4] = s.sets[4]&^(1<<7) | 1
	}
	s.domStar, s.dowStar = strings.HasPrefix(parts[2], "*"), strings.HasPrefix(parts[4], "*")
	return s, nil
}

func parseField(s string, f cronField) (uint64, error) {
	var set uint64
	for _, part := range strings.Split(s, ",") {
		rng, stepStr, hasStep := strings.Cut(part, "/")
		step := 1
		if hasStep {
			n, err := strconv.Atoi(stepStr)
			if err != nil || n < 1 {
				return 0, errors.New("invalid step " + strconv.Quote(stepStr))
			}
			step = n
		}
		lo, hi := f.min, f.max
		if rng != "*" {
			a, b, isRange := strings.Cut(rng, "-")
			var err error
			if lo, err = cronValue(a, f); err != nil {
				return 0, err
			}
			hi = lo
			if isRange {
				if hi, err = cronValue(b, f); err != nil {
					return 0, err
				}
			} else if hasStep {
				hi = f.max // "5/15" means 5-max/15
			}
			if lo > hi {
				return 0, errors.New("range " + strconv.Quote(rng) + " is reversed")
			}
		}
		for v := lo; v <= hi; v += step {
			set |= 1 << v
		}
	}
	return set, nil
}

func cronValue(s string, f cronField) (int, error) {
	for i, n := range f.names {
		if strings.EqualFold(s, n) {
			return f.min + i, nil
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < f.min || n > f.max {
		return 0, errors.New(strconv.Quote(s) + " is not in " + strconv.Itoa(f.min) + "-" + strconv.Itoa(f.max))
	}
	return n, nil
}

func describe(set uint64, f cronField) string {
	max := f.max
	if f.name == "day of week" {
		max = 6
	}
	if bits.OnesCount64(set) == max-f.min+1 {
		return "every " + f.name
	}
	var vals []string
	for v := f.min; v <= max; v++ {
		if set&(1<<v) != 0 {
			s := strconv.Itoa(v)
			if f.names != nil {
				s = f.names[v-f.min]
			}
			vals = append(vals, s)
		}
	}
	return strings.Join(vals, ", ")
}

// next returns the first matching minute after t, skipping whole months, days and hours that cannot match.
func (s cronSpec) next(t time.Time) (time.Time, error) {
	loc := t.Location()
	limit := t.AddDate(5, 0, 0)
	t = t.Truncate(time.Minute).Add(time.Minute)
	for t.Before(limit) {
		switch {
		case s.sets[3]&(1<<int(t.Month())) == 0:
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, loc)
		case !s.dayMatches(t):
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
		case s.sets[1]&(1<<t.Hour()) == 0:
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, loc)
		case s.sets[0]&(1<<t.Minute()) == 0:
			t = t.Add(time.Minute)
		default:
			return t, nil
		}
	}
	return time.Time{}, errors.New("no run time in the next 5 years")
}

// dayMatches follows cron: when both day fields are restricted, either one matching is enough.
func (s cronSpec) dayMatches(t time.Time) bool {
	dom, dow := s.sets[2]&(1<<t.Day()) != 0, s.sets[4]&(1<<int(t.Weekday())) != 0
	if s.domStar || s.dowStar {
		return dom && dow
	}
	return dom || dow
}
