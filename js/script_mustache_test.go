package js

import (
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
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
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "foobar", v.String())
			},
		},
		{
			"render with fake and function",
			func(t *testing.T) {
				s, err := New("test",
					`import { render } from 'mokapi/mustache';
						 import { fake } from 'mokapi/faker'

						 export default function() {
							const scope = {
								firstname: fake({
									type: 'string',
									format: '{firstname}'
								}),
								calc: () => ( 3 + 4 )
							};
						
							return render("{{firstname}} has {{calc}} apples", scope);
						}`,
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "Markus has 7 apples", v.String())
			},
		},
		{
			"using deprecated module",
			func(t *testing.T) {
				s, err := New("test",
					`import {render} from 'mustache';
						 export default function() {
						 	return render('foo{{x}}', {x: 'bar'})
						}`,
					host, static.JsConfig{})
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
			gofakeit.Seed(11)
			tc.f(t)
		})
	}
}
