package urlpath

import "testing"

type data struct {
	pat    string
	lit    string
	result bool
}

var tests = [...]data{
	{"foobar", "foobar", true},
	{"foo", "foobar", false},
	{"foo*", "foobar", true},
	{"foo*ar", "foobar", true},
	{"foo*ar", "foo/ar", false},
	{"foo/*ar", "foo/ar", true},
	{"*.jpg", "test.jpg", true},
	{"**.jpg", "/a/b/c/test.jpg", true},
	{"/a/**/test.jpg", "/a/b/c/test.jpg", true},
	{"/a/**/t*.jpg", "/a/b/c/test.jpg", true},
	{"/a/**/x*.jpg", "/a/b/c/test.jpg", false},
}

func TestSimple(t *testing.T) {
	for _, test := range tests {
		if b := Match(test.pat, test.lit); b != test.result {
			t.Errorf("Expected %v, got %v: pattern %v, value: %v", test.result, b, test.pat, test.lit)
		}
	}
}
