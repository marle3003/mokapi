package js

import (
	r "github.com/stretchr/testify/require"
	"testing"
)

func Test_Mustache(t *testing.T) {
	host := &testHost{}
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"render",
			func(t *testing.T) {
				s, err := New("test",
					`import {render} from 'mokapi/mustache';
						 export default function() {
						 	return render('foo{{x}}', {x: 'bar'})
						}`,
					host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "foobar", v.String())
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}
