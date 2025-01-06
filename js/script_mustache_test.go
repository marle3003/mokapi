package js_test

import (
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
)

func Test_Mustache(t *testing.T) {
	host := &enginetest.Host{}
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "render",
			test: func(t *testing.T) {
				s, err := jstest.New(jstest.WithSource(
					`import {render} from 'mokapi/mustache';
						 export default function() {
						 	return render('foo{{x}}', {x: 'bar'})
						}`),
					js.WithHost(host))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "foobar", v.String())
			},
		},
		{
			name: "render with fake and function",
			test: func(t *testing.T) {
				s, err := jstest.New(jstest.WithSource(
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
						}`),
					js.WithHost(host))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "Markus has 7 apples", v.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			tc.test(t)
		})
	}
}
