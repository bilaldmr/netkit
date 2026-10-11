package schema

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Google rich result rules on top of the schema.org vocabulary, which has no
// required properties or value ranges. Google publishes these only as prose,
// so the table is maintained by hand from
// https://developers.google.com/search/docs/appearance/structured-data/search-gallery
// A rule applies to the type and its subtypes (Article covers NewsArticle, BlogPosting).
// "a|b" means one of a or b. Recommended is only checked on top-level and @graph entities.
type typeRule struct {
	Required     []string
	RootRequired []string // required only when not nested (itemReviewed)
	Recommended  []string
	Exact        bool // skip subtypes, ListItem must not cover HowToStep
}

var googleRules = map[string]typeRule{
	"Article":                {Recommended: []string{"headline", "image", "author", "datePublished", "dateModified"}},
	"BreadcrumbList":         {Required: []string{"itemListElement"}},
	"ListItem":               {Required: []string{"position", "name|item"}, Exact: true},
	"HowToStep":              {Required: []string{"text|itemListElement"}},
	"Product":                {Required: []string{"name", "offers|review|aggregateRating"}, Recommended: []string{"image", "description", "brand", "sku|gtin|gtin8|gtin12|gtin13|gtin14|mpn", "offers", "review", "aggregateRating"}},
	"Offer":                  {Required: []string{"price|priceSpecification", "priceCurrency|priceSpecification"}, Recommended: []string{"availability", "url"}},
	"AggregateOffer":         {Required: []string{"lowPrice", "priceCurrency"}, Recommended: []string{"highPrice", "offerCount"}},
	"Review":                 {Required: []string{"author", "reviewRating"}, RootRequired: []string{"itemReviewed"}, Recommended: []string{"datePublished"}},
	"Rating":                 {Required: []string{"ratingValue"}},
	"AggregateRating":        {Required: []string{"ratingCount|reviewCount"}, RootRequired: []string{"itemReviewed"}},
	"Recipe":                 {Required: []string{"name", "image"}, Recommended: []string{"author", "datePublished", "description", "recipeIngredient", "recipeInstructions", "recipeYield", "prepTime", "cookTime", "totalTime", "aggregateRating", "nutrition", "keywords", "recipeCategory", "recipeCuisine"}},
	"FAQPage":                {Required: []string{"mainEntity"}},
	"QAPage":                 {Required: []string{"mainEntity"}},
	"Question":               {Required: []string{"name", "acceptedAnswer|suggestedAnswer"}},
	"Answer":                 {Required: []string{"text"}},
	"Event":                  {Required: []string{"name", "startDate", "location"}, Recommended: []string{"description", "endDate", "eventStatus", "eventAttendanceMode", "image", "offers", "organizer", "performer"}},
	"LocalBusiness":          {Required: []string{"name", "address"}, Recommended: []string{"telephone", "url", "geo", "openingHoursSpecification", "priceRange", "image"}},
	"Organization":           {Recommended: []string{"name", "url", "logo", "sameAs"}},
	"JobPosting":             {Required: []string{"title", "description", "datePosted", "hiringOrganization", "jobLocation|applicantLocationRequirements"}, Recommended: []string{"validThrough", "employmentType", "baseSalary", "identifier"}},
	"VideoObject":            {Required: []string{"name", "thumbnailUrl", "uploadDate"}, Recommended: []string{"description", "duration", "contentUrl|embedUrl"}},
	"SoftwareApplication":    {Required: []string{"name", "offers", "aggregateRating|review"}, Recommended: []string{"applicationCategory", "operatingSystem"}},
	"Movie":                  {Required: []string{"name", "image"}, Recommended: []string{"dateCreated", "director", "aggregateRating", "review"}},
	"Dataset":                {Required: []string{"name", "description"}, Recommended: []string{"license", "creator", "url"}},
	"ProfilePage":            {Required: []string{"mainEntity"}},
	"DiscussionForumPosting": {Required: []string{"author", "datePublished", "text|image|video"}},
	"Course":                 {Required: []string{"name", "description"}, Recommended: []string{"provider"}},
	"PostalAddress":          {Recommended: []string{"streetAddress", "addressLocality", "postalCode", "addressCountry"}},
}

// Value checks keyed by property, so they run on every type that uses it.
var (
	dateProps     = set("datePublished", "dateModified", "dateCreated", "uploadDate", "datePosted", "validThrough", "validFrom", "startDate", "endDate", "priceValidUntil")
	durationProps = set("cookTime", "prepTime", "totalTime", "performTime", "duration")
	priceProps    = set("price", "lowPrice", "highPrice")
	countProps    = set("ratingCount", "reviewCount", "offerCount")
	enumProps     = set("availability", "itemCondition", "eventStatus", "eventAttendanceMode")

	durationRe  = regexp.MustCompile(`^P(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T(\d+H)?(\d+M)?(\d+(\.\d+)?S)?)?$`)
	currencyRe  = regexp.MustCompile(`^[A-Z]{3}$`)
	dateLayouts = []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02T15:04Z07:00", "2006-01-02"}
)

func (v *Vocab) rules(n map[string]any, types []string, path string, b *Block, root bool) {
	add := func(level, prop, msg string) {
		p := path
		if prop != "" {
			p += "." + prop
		}
		b.Issues = append(b.Issues, Issue{level, p, msg, "google"})
	}

	seen := map[string]bool{}
	need := func(level, prefix string, alts []string) {
		for _, alt := range alts {
			if seen[alt] {
				continue
			}
			seen[alt] = true
			opts := strings.Split(alt, "|")
			switch {
			case !anyPresent(n, opts) && len(opts) > 1:
				add(level, "", "one of "+strings.Join(opts, ", ")+" is "+prefix)
			case !anyPresent(n, opts):
				add(level, "", "missing "+prefix+" "+alt)
			case len(opts) > 1:
				// Google ignores an alternative with errors, so a broken
				// aggregateRating does not satisfy offers|review|aggregateRating.
				var valid, broken []string
				for _, o := range opts {
					if !anyPresent(n, []string{o}) {
						continue
					}
					if hasErrors(b, path+"."+o) {
						broken = append(broken, o)
					} else {
						valid = append(valid, o)
					}
				}
				if len(valid) == 0 {
					add(level, "", "one of "+strings.Join(opts, ", ")+" is "+prefix+", "+strings.Join(broken, ", ")+" has errors")
				}
			}
		}
	}
	for name, r := range googleRules {
		if r.Exact && !slices.Contains(types, name) || !r.Exact && !v.isAny(types, name) {
			continue
		}
		need("error", "required", r.Required)
		if root {
			need("error", "required", r.RootRequired)
			need("warning", "recommended", r.Recommended)
		}
	}

	for k, val := range n {
		s, isStr := val.(string)
		switch {
		case dateProps[k] && isStr && !isDate(s):
			add("error", k, fmt.Sprintf("%q is not an ISO 8601 date", s))
		case durationProps[k] && isStr && (!durationRe.MatchString(s) || s == "P" || strings.HasSuffix(s, "T")):
			add("error", k, fmt.Sprintf("%q is not an ISO 8601 duration like PT1H30M", s))
		case k == "priceCurrency" && isStr && !currencyRe.MatchString(s):
			add("error", k, fmt.Sprintf("%q is not an ISO 4217 currency code like EUR", s))
		case priceProps[k]:
			if f, ok := num(val); !ok || f < 0 {
				add("error", k, fmt.Sprintf("%v must be a number without currency symbols, with a dot as decimal separator", val))
			}
		case countProps[k]:
			if f, ok := num(val); !ok || f <= 0 {
				add("error", k, k+" must be a positive number")
			}
		case enumProps[k] && isStr && !v.Members[term(s)]:
			add("error", k, fmt.Sprintf("%q is not a schema.org enumeration value like https://schema.org/InStock", s))
		}
	}

	if v.isAny(types, "Rating") && n["ratingValue"] != nil {
		value, ok := num(n["ratingValue"])
		best, hasBest := num(n["bestRating"])
		worst, hasWorst := num(n["worstRating"])
		if !hasBest {
			best = 5
		}
		if !hasWorst {
			worst = 1
		}
		switch {
		case !ok:
			add("error", "ratingValue", "ratingValue must be a number")
		case best <= worst:
			add("error", "bestRating", fmt.Sprintf("bestRating (%g) must be greater than worstRating (%g)", best, worst))
		case value < worst || value > best:
			add("error", "ratingValue", fmt.Sprintf("ratingValue %g is out of range %g to %g", value, worst, best))
		}
	}
}

func (v *Vocab) isAny(types []string, ancestor string) bool {
	for _, t := range types {
		if v.isA(t, ancestor, 0) {
			return true
		}
	}
	return false
}

// hasErrors reports an error issue at prop or below it.
func hasErrors(b *Block, prop string) bool {
	for _, i := range b.Issues {
		if i.Level == "error" && (i.Path == prop || strings.HasPrefix(i.Path, prop+".") || strings.HasPrefix(i.Path, prop+"[")) {
			return true
		}
	}
	return false
}

// anyPresent treats null, "" and [] as missing.
func anyPresent(n map[string]any, keys []string) bool {
	for _, k := range keys {
		switch x := n[k].(type) {
		case nil:
		case string:
			if strings.TrimSpace(x) != "" {
				return true
			}
		case []any:
			if len(x) > 0 {
				return true
			}
		default:
			return true
		}
	}
	return false
}

func isDate(s string) bool {
	for _, l := range dateLayouts {
		if _, err := time.Parse(l, s); err == nil {
			return true
		}
	}
	return false
}

// num reads a JSON number or numeric string.
func num(x any) (float64, bool) {
	switch t := x.(type) {
	case float64:
		return t, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	}
	return 0, false
}

func set(keys ...string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}
