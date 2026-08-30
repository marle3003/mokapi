package asyncapi3_test

import (
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, jsonString, yamlString string, err error)
	}{
		{
			name: "server with global ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseServer("foo", &asyncapi3.ServerRef{
					Reference: dynamic.Reference[*asyncapi3.ServerRef]{
						Ref: "/foo",
					},
					Value: &asyncapi3.Server{Host: "localhost"},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","servers":{"foo":{"$ref":"#/components/servers/foo"}},"components":{"servers":{"foo":{"host":"localhost"}}}}`, jsonString)
			},
		},
		{
			name: "server local ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseServer("foo", &asyncapi3.ServerRef{
					Reference: dynamic.Reference[*asyncapi3.ServerRef]{
						Ref: "#/components/servers/foo",
					},
					Value: &asyncapi3.Server{Host: "localhost"},
				}),
				asyncapi3test.WithComponentServer("foo", &asyncapi3.Server{Host: "localhost"}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","servers":{"foo":{"$ref":"#/components/servers/foo"}},"components":{"servers":{"foo":{"host":"localhost"}}}}`, jsonString)
			},
		},
		{
			name: "server no ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseServer("foo", &asyncapi3.ServerRef{
					Value: &asyncapi3.Server{Host: "localhost"},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","servers":{"foo":{"host":"localhost"}}}`, jsonString)
			},
		},
		{
			name: "channel with global ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Reference: dynamic.Reference[*asyncapi3.ChannelRef]{
						Ref: "/foo",
					},
					Value: &asyncapi3.Channel{Description: "desc"},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"$ref":"#/components/channels/foo"}},"components":{"channels":{"foo":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel local ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Reference: dynamic.Reference[*asyncapi3.ChannelRef]{
						Ref: "#/components/servers/foo",
					},
					Value: &asyncapi3.Channel{Description: "desc"},
				}),
				asyncapi3test.WithComponentChannel("foo", &asyncapi3.Channel{Description: "desc"}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"$ref":"#/components/channels/foo"}},"components":{"channels":{"foo":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel no ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{Description: "desc"},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"description":"desc"}}}`, jsonString)
			},
		},
		{
			name: "channel with server ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseServer("bar", &asyncapi3.ServerRef{
					Value: &asyncapi3.Server{Host: "localhost"},
				}),
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Servers: []*asyncapi3.ServerRef{
							{
								Reference: dynamic.Reference[*asyncapi3.ServerRef]{
									Ref: "#/servers/bar",
								},
								Value: &asyncapi3.Server{Host: "localhost"},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","servers":{"bar":{"host":"localhost"}},"channels":{"foo":{"servers":[{"$ref":"#/servers/bar"}]}}}`, jsonString)
			},
		},
		{
			name: "channel with global message ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Messages: map[string]*asyncapi3.MessageRef{
							"msg": {
								Reference: dynamic.Reference[*asyncapi3.MessageRef]{
									Ref: "/bar",
								},
								Value: &asyncapi3.Message{Description: "desc"},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"messages":{"msg":{"$ref":"#/components/messages/bar"}}}},"components":{"messages":{"bar":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel with local message ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Messages: map[string]*asyncapi3.MessageRef{
							"msg": {
								Reference: dynamic.Reference[*asyncapi3.MessageRef]{
									Ref: "#/components/messages/bar",
								},
								Value: &asyncapi3.Message{Description: "desc"},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"messages":{"msg":{"$ref":"#/components/messages/bar"}}}},"components":{"messages":{"bar":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel with global correlationId",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Messages: map[string]*asyncapi3.MessageRef{
							"msg": {
								Reference: dynamic.Reference[*asyncapi3.MessageRef]{
									Ref: "#/components/messages/bar",
								},
								Value: &asyncapi3.Message{
									CorrelationId: &asyncapi3.CorrelationIdRef{
										Reference: dynamic.Reference[*asyncapi3.CorrelationIdRef]{
											Ref: "/cor1",
										},
										Value: &asyncapi3.CorrelationId{
											Description: "desc",
										},
									},
								},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"messages":{"msg":{"$ref":"#/components/messages/bar"}}}},"components":{"messages":{"bar":{"correlationId":{"$ref":"#/components/correlationIds/cor1"}}},"correlationIds":{"cor1":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel with correlationId",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Messages: map[string]*asyncapi3.MessageRef{
							"msg": {
								Reference: dynamic.Reference[*asyncapi3.MessageRef]{
									Ref: "#/components/messages/bar",
								},
								Value: &asyncapi3.Message{
									CorrelationId: &asyncapi3.CorrelationIdRef{
										Reference: dynamic.Reference[*asyncapi3.CorrelationIdRef]{
											Ref: "#/components/correlationIds/cor1",
										},
										Value: &asyncapi3.CorrelationId{
											Description: "desc",
										},
									},
								},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"messages":{"msg":{"$ref":"#/components/messages/bar"}}}},"components":{"messages":{"bar":{"correlationId":{"$ref":"#/components/correlationIds/cor1"}}},"correlationIds":{"cor1":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel with message trait",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Messages: map[string]*asyncapi3.MessageRef{
							"msg": {
								Reference: dynamic.Reference[*asyncapi3.MessageRef]{
									Ref: "#/components/messages/bar",
								},
								Value: &asyncapi3.Message{
									Traits: []*asyncapi3.MessageTraitRef{
										{
											Reference: dynamic.Reference[*asyncapi3.MessageTraitRef]{
												Ref: "#/components/messageTraits/trait",
											},
											Value: &asyncapi3.MessageTrait{
												Description: "desc",
											},
										},
									},
								},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"messages":{"msg":{"$ref":"#/components/messages/bar"}}}},"components":{"messages":{"bar":{"traits":[{"$ref":"#/components/messageTraits/trait"}]}},"messageTraits":{"trait":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel with external docs",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Messages: map[string]*asyncapi3.MessageRef{
							"msg": {
								Reference: dynamic.Reference[*asyncapi3.MessageRef]{
									Ref: "#/components/messages/bar",
								},
								Value: &asyncapi3.Message{
									ExternalDocs: []*asyncapi3.ExternalDocRef{
										{
											Reference: dynamic.Reference[*asyncapi3.ExternalDocRef]{
												Ref: "#/components/messageTraits/ex",
											},
											Value: &asyncapi3.ExternalDoc{
												Description: "desc",
											},
										},
									},
								},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"messages":{"msg":{"$ref":"#/components/messages/bar"}}}},"components":{"messages":{"bar":{"externalDocs":[{"$ref":"#/components/externalDocs/ex"}]}},"externalDocs":{"ex":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "channel with parameter",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseChannel("foo", &asyncapi3.ChannelRef{
					Value: &asyncapi3.Channel{
						Parameters: map[string]*asyncapi3.ParameterRef{
							"param": {
								Reference: dynamic.Reference[*asyncapi3.ParameterRef]{
									Ref: "#/components/parameters/param",
								},
								Value: &asyncapi3.Parameter{
									Description: "desc",
								},
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"foo":{"parameters":{"param":{"$ref":"#/components/parameters/param"}}}},"components":{"parameters":{"param":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "operation with global channel",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseOperation("foo", &asyncapi3.OperationRef{
					Value: &asyncapi3.Operation{
						Channel: &asyncapi3.ChannelRef{
							Reference: dynamic.Reference[*asyncapi3.ChannelRef]{
								Ref: "#/components/messages/bar",
							},
							Value: &asyncapi3.Channel{
								Description: "desc",
							},
						},
					},
				}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","operations":{"foo":{"channel":{"$ref":"#/components/channels/bar"}}},"components":{"channels":{"bar":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "operation with local channel",
			cfg: func() *asyncapi3.Config {
				ch := &asyncapi3.ChannelRef{
					Reference: dynamic.Reference[*asyncapi3.ChannelRef]{
						Ref: "#/components/channels/bar",
					},
					Value: &asyncapi3.Channel{
						Description: "desc",
					},
				}

				cfg := asyncapi3test.NewConfig(
					asyncapi3test.UseChannel("ch", ch),
					asyncapi3test.UseOperation("foo", &asyncapi3.OperationRef{
						Value: &asyncapi3.Operation{
							Channel: ch,
						},
					}),
					asyncapi3test.WithComponentChannel("bar", ch.Value),
				)
				return cfg
			}(),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","channels":{"ch":{"$ref":"#/components/channels/bar"}},"operations":{"foo":{"channel":{"$ref":"#/channels/ch"}}},"components":{"channels":{"bar":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "operation with message schema ref",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.UseOperation("foo", &asyncapi3.OperationRef{
					Value: &asyncapi3.Operation{
						Messages: []*asyncapi3.MessageRef{
							{
								Value: asyncapi3test.NewMessage(
									asyncapi3test.UsePayload(
										&asyncapi3.SchemaRef{
											Reference: dynamic.Reference[*asyncapi3.SchemaRef]{
												Ref: "#/components/schemas/s1",
											},
											Value: schematest.New("string"),
										},
									),
								),
							},
						},
					},
				}),
				asyncapi3test.WithComponentSchema("s1", schematest.New("string")),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"asyncapi":"3.0.0","info":{"title":"test","version":"1.0"},"defaultContentType":"application/json","operations":{"foo":{"messages":[{"payload":{"$ref":"#/components/schemas/s1"}}]}},"components":{"schemas":{"s1":{"type":"string"}}}}`, jsonString)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := asyncapi3.Export{}
			jsonResult, err := e.ToJSON(tc.cfg)
			tc.test(t, string(jsonResult), "", err)
		})
	}
}
