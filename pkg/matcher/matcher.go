package matcher

import (
	"fmt"
	"regexp"
	"strings"
)

// Sample usage
//
// pattern := `/get/{var1}/{var2}/?a={var3}`
// pm := NewPatternMatching(pattern)
//
// pm.Match("/get/130/20/?a=99")
//
// res := pm.ApplyResults("/new?v1={var1}&v2={var2}")
// fmt.Println(res)
type Matcher struct {
	re *regexp.Regexp
}

func NewPatternMatching(pattern string) *Matcher {
	pattern = strings.ReplaceAll(pattern, "?", `\?`)
	pattern = strings.ReplaceAll(pattern, "{", "(?P<")
	pattern = strings.ReplaceAll(pattern, "/", `\/?`)
	pattern = strings.ReplaceAll(pattern, "}", ">[^/]+)?")
	fmt.Println(pattern + "\n")

	pm := &Matcher{
		re: regexp.MustCompile(pattern),
	}
	return pm
}

func (p *Matcher) Transform(path string, pattern string) string {
	match := p.re.FindStringSubmatch(path)
	results := make(map[string]string)
	for i, name := range p.re.SubexpNames() {
		if i != 0 && name != "" && i < len(match) {
			results[name] = match[i]
		}
	}

	out := pattern
	for k, v := range results {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	return out
}
