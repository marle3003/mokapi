package js_test

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"mokapi/schema/json/generator"
	"testing"
)

func TestScript_Faker(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "fake string",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: 'string' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq ", v.String())
			},
		},
		{
			name: "fake string or number",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: ['string', 'number'] })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, 1.4470162467581058e+308, v.Export())
			},
		},
		{
			name: "invalid type for type",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: {} })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.EqualError(t, err, "unexpected type for 'type': Object at mokapi/js/faker.(*Faker).Fake-fm (native)")
			},
		},
		{
			name: "fake exclusiveMinimum",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: 'integer', exclusiveMinimum: 3, maximum: 10 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "5", v.String())
			},
		},
		{
			name: "fake exclusiveMinimum but wrong type",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: 'integer', exclusiveMinimum: 'str', maximum: 10 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.EqualError(t, err, "unexpected type for 'exclusiveMinimum': String at mokapi/js/faker.(*Faker).Fake-fm (native)")
			},
		},
		{
			name: "fake exclusiveMaximum",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: 'integer', exclusiveMaximum: 3, minimum: 2 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "2", v.String())
			},
		},
		{
			name: "fake exclusiveMaximum but wrong type",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(`import faker from 'mokapi/faker'
						 export default function() {
						 	return faker.fake({ type: 'integer', exclusiveMaximum: 'str', minimum: 10 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.EqualError(t, err, "unexpected type for 'exclusiveMaximum': String at mokapi/js/faker.(*Faker).Fake-fm (native)")
			},
		},
		{
			name: "fake enum",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { fake } from 'mokapi/faker'
						 export default function() {
						 	return fake({ type: 'string', enum: ['foo', 'bar'] })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.String())
			},
		},
		{
			name: "invalid format",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { fake } from 'mokapi/faker'
						 export default function() {
						 	return fake("")
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.EqualError(t, err, "expect object parameter but got: String at mokapi/js/faker.(*Faker).Fake-fm (native)")
			},
		},
		{
			name: "object with properties",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import { fake } from 'mokapi/faker'
						 export default function() {
						 	return fake({ type: 'object', required: ['id'], properties: { id: { type: 'integer', format: 'int64' } } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				b, err := json.Marshal(v)
				r.NoError(t, err)
				r.Equal(t, `{"id":98266}`, string(b))
			},
		},
		{
			name: "find node",
			test: func(t *testing.T, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *common.FakerTree {
					return common.NewFakerTree(generator.FindByName(name))
				}

				s, err := jstest.New(jstest.WithSource(
					`import { fake, findByName } from 'mokapi/faker'
						 export default function() {
						 	let root = findByName('')
							root.append({
								name: 'foo',
								fake: (r) => {
									return 'bar'
								}
							})
							return fake({ type: 'object', properties: { foo: { type: 'string'} } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{
					"foo": "bar",
				}, v.Export())
				r.NoError(t, err)
				n := generator.FindByName("")
				err = n.Remove("foo")
				r.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			tc.test(t, &enginetest.Host{})
		})
	}
}
