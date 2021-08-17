package internal

import "regexp"

var (
	refTagsRe = regexp.MustCompile(`refs/tags/(.*)-(\d\.\d\.\d.*)`)
	gitRefRe  = regexp.MustCompile(`(.*)-(\d\.\d\.\d.*)`)
)

type feature struct {
	Name    string
	Version string
}

func (f feature) isEmpty() bool {
	return f == (feature{})
}

func parseToFeature(target string, re *regexp.Regexp) feature {
	matches := re.FindAllStringSubmatch(target, -1)
	if len(matches) == 0 {
		return feature{}
	}
	match := matches[0]
	return feature{Name: match[1], Version: match[2]}
}
