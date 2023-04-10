package proxy

import (
	"regexp"
	"strconv"
	"strings"
)

/*
// Matcher for rewrite rules.
// Sample usage:
//
//   - upstream: https://httpbin.org/
//     path: /mnt/
//     middlewares:
//       rewrite:
//         pattern: /get/(.+)
//         target: /get?id=$1
//
// This will rewrirte this call
//
//	curl -s "http://localhost/mnt/get/1"
//
// using the target
//
//	https://httpbin.org/get?id=1
*/
type rewriter struct {
	regexp    *regexp.Regexp
	transform string
}

func newRewriter(pattern string, transform string) *rewriter {
	pm := &rewriter{
		regexp:    regexp.MustCompile(pattern),
		transform: transform,
	}
	return pm
}

func (p *rewriter) rewrite(input string) string {

	groups := p.regexp.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return input
	}
	values := groups[0][1:]

	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	replacer := strings.NewReplacer(replace...)
	return replacer.Replace(p.transform)
}
