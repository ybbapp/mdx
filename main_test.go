package main

import "testing"

func TestParseArgs(t *testing.T) {
	cases := []struct {
		args       []string
		mode, path string
		bad        bool
	}{
		{[]string{"example.mdx"}, "view", "example.mdx", false},
		{[]string{"cat", "example.mdx"}, "cat", "example.mdx", false},
		{[]string{"./cat"}, "view", "./cat", false},
		{[]string{"--help"}, "help", "", false},
		{nil, "", "", true}, {[]string{"cat"}, "", "", true}, {[]string{"lens", "example.mdx"}, "", "", true},
	}
	for _, tc := range cases {
		mode, path, err := parseArgs(tc.args)
		if (err != nil) != tc.bad || mode != tc.mode || path != tc.path {
			t.Errorf("%v: %q %q %v", tc.args, mode, path, err)
		}
	}
}
