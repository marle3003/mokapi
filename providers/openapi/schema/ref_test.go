package schema_test

/*func TestRef_HasProperties(t *testing.T) {
	r := &schema.Ref{}
	require.False(t, r.HasProperties())

	r.Value = &schema.Schema{}
	require.False(t, r.HasProperties())

	r.Value.Properties = &schema.Schemas{}
	require.False(t, r.HasProperties())

	r.Value.Properties.Set("foo", nil)
	require.True(t, r.HasProperties())
}

func TestRef_String(t *testing.T) {
	r := &schema.Ref{}
	require.Equal(t, "no schema defined", r.String())

	r = &schema.Ref{Reference: dynamic.Reference{Ref: "foo"}}
	require.Equal(t, "unresolved schema foo", r.String())

	r.Value = &schema.Schema{}
	require.Equal(t, "", r.String())

	r.Value.Type = jsonSchema.Types{"number"}
	require.Equal(t, "schema type=number", r.String())
}

func TestRef_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Ref is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				var r *schema.Ref
				err := r.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: r}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "with reference",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: schematest.New("number")}
					return cfg, nil
				})
				r := &schema.Ref{Reference: dynamic.Reference{Ref: "foo.yml"}}
				err := r.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: r}, reader)
				require.NoError(t, err)
				require.NotNil(t, r.Value)
				require.Equal(t, "number", r.Value.Type.String())
			},
		},
		{
			name: "with reference but error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				r := &schema.Ref{Reference: dynamic.Reference{Ref: "foo.yml"}}
				err := r.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: r}, reader)
				require.EqualError(t, err, "parse schema failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "value is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
				})
				r := &schema.Ref{}
				err := r.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: r}, reader)
				require.NoError(t, err)
				require.Nil(t, r.Value)
			},
		},
		{
			name: "with value",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					return &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}, nil
				})
				r := &schema.Ref{Value: schematest.New("integer")}
				err := r.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: r}, reader)
				require.NoError(t, err)
				require.NotNil(t, r.Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestRef_UnmarshalJSON(t *testing.T) {
	for _, testcase := range []struct {
		name string
		s    string
		fn   func(t *testing.T, r *schema.Ref)
	}{
		{
			name: "ref",
			s:    `{ "$ref": "#/components/schema/Foo" }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "#/components/schema/Foo", r.Ref)
			},
		},
	} {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			r := &schema.Ref{}
			err := json.Unmarshal([]byte(test.s), r)
			require.NoError(t, err)
			test.fn(t, r)
		})
	}
}

func TestRef_UnmarshalYAML(t *testing.T) {
	for _, testcase := range []struct {
		name string
		s    string
		fn   func(t *testing.T, r *schema.Ref)
	}{
		{
			name: "ref",
			s:    "$ref: '#/components/schema/Foo'",
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "#/components/schema/Foo", r.Ref)
			},
		},
	} {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			r := &schema.Ref{}
			err := yaml.Unmarshal([]byte(test.s), r)
			require.NoError(t, err)
			test.fn(t, r)
		})
	}
}

func TestRef_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Ref
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "empty type",
			s:    &schema.Ref{},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "{}", s)
			},
		},
		{
			name: "with type",
			s:    &schema.Ref{Value: schematest.New("string")},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":"string"}`, s)
			},
		},
		{
			name: "with properties",
			s:    schematest.NewRef("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":"object","properties":{"foo":{"type":"string"}}}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.s)
			tc.test(t, string(b), err)
		})
	}
}*/
