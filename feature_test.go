package main

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_parseToFeature(t *testing.T) {
	testCases := []struct {
		target   string
		desc     string
		re       *regexp.Regexp
		expected feature
	}{
		{
			target: "d6602ec5194c87b0fc87103ca4d67251c76f233a	refs/tags/feature-1.1.0",
			desc:     "refTagsRe",
			re:       refTagsRe,
			expected: feature{Name: "feature", Version: "1.1.0"},
		},
		{
			target:   "feature-1.1.0",
			desc:     "gitRefRe",
			re:       gitRefRe,
			expected: feature{Name: "feature", Version: "1.1.0"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := parseToFeature(tC.target, tC.re)
			if !reflect.DeepEqual(got, tC.expected) {
				t.Errorf("Expected %+v, want %+v", tC.expected, got)
			}
		})
	}
}
