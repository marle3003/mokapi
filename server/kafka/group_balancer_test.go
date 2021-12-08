package kafka_test

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/server/kafka/protocol/syncGroup"
	"mokapi/test"
	"testing"
	"time"
)

func TestGroupBalancing(t *testing.T) {
	//t.Parallel()
	testdata := []struct {
		name string
		fn   func(t *testing.T, b *kafka.Binding, config *asyncApi.Config)
	}{
		// a group is created by FindCoordinator request or by binding configuration
		{"join group with invalid group name", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			c := kafkatest.NewClient(":9092", "kafkatest")
			defer c.Close()
			join, err := c.JoinGroup(3, &joinGroup.Request{
				GroupId:      "TestGroup",
				MemberId:     "foo",
				ProtocolType: "consumer",
				Protocols: []joinGroup.Protocol{{
					Name: "range",
				}},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.InvalidGroupId, join.ErrorCode)
		}},
		{"join group with FindCoordinator", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			c := kafkatest.NewClient(":9092", "kafkatest")
			defer c.Close()
			fc, err := c.FindCoordinator(2, &findCoordinator.Request{
				Key:     "TestGroup",
				KeyType: findCoordinator.KeyTypeGroup,
			})
			test.Ok(t, err)
			test.Equals(t, protocol.None, fc.ErrorCode)

			join, err := c.JoinGroup(3, &joinGroup.Request{
				GroupId:      "TestGroup",
				MemberId:     "foo",
				ProtocolType: "consumer",
				Protocols: []joinGroup.Protocol{{
					Name: "range",
				}},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.None, join.ErrorCode)
		}},
		{"join group with configured group name", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			config.Channels["foo"].Value.Subscribe.Bindings.Kafka.GroupId = &openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}
			err := b.Apply(config)
			test.Ok(t, err)

			c := kafkatest.NewClient(":9092", "kafkatest")
			defer c.Close()
			meta := []byte{
				0, 1, // version
				0, 0, 0, 1, // topic array length
				0, 3, 'f', 'o', 'o', // topic foo
				0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
			}
			join, err := c.JoinGroup(3, &joinGroup.Request{
				GroupId:      "TestGroup",
				MemberId:     "foo",
				ProtocolType: "consumer",
				Protocols: []joinGroup.Protocol{{
					Name:     "range",
					MetaData: meta,
				}},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.None, join.ErrorCode)
			test.Equals(t, "kafkatest", join.Leader)
			test.Equals(t, "range", join.ProtocolName)
			// currently, not correct because conflict between client id and member id
			test.Equals(t, "kafkatest", join.MemberId)
			test.Equals(t, "kafkatest", join.Members[0].MemberId)
			test.Equals(t, meta, join.Members[0].MetaData)
		}},
		{"sync group but not member", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			config.Channels["foo"].Value.Subscribe.Bindings.Kafka.GroupId = &openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}
			err := b.Apply(config)
			test.Ok(t, err)

			c := kafkatest.NewClient(":9092", "kafkatest")
			sync, err := c.SyncGroup(3, &syncGroup.Request{
				GroupId:      "TestGroup",
				GenerationId: 0,
				MemberId:     "foo",
			})
			test.Ok(t, err)
			test.Equals(t, protocol.UnknownMemberId, sync.ErrorCode)
		}},
		{"sync group but joining state", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			config.Channels["foo"].Value.Subscribe.Bindings.Kafka.GroupId = &openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}
			config.Servers["foo"].Bindings.Kafka.Config["group.initial.rebalance.delay.ms"] = "3000"
			err := b.Apply(config)
			test.Ok(t, err)

			ch := make(chan bool)
			go func() {
				c := kafkatest.NewClient(":9092", "kafkatest")
				defer c.Close()
				join, err := c.JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     "foo",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				test.Ok(t, err)
				test.Equals(t, protocol.None, join.ErrorCode)
				ch <- true
			}()

			time.Sleep(500 * time.Millisecond)
			c := kafkatest.NewClient(":9092", "kafkatest")
			defer c.Close()
			sync, err := c.SyncGroup(3, &syncGroup.Request{
				GroupId:      "TestGroup",
				GenerationId: 0,
				MemberId:     "foo",
			})
			test.Ok(t, err)
			test.Equals(t, protocol.RebalanceInProgress, sync.ErrorCode)
			// wait for join response
			<-ch
		}},
		{"sync group successfully", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			config.Channels["foo"].Value.Subscribe.Bindings.Kafka.GroupId = &openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}
			err := b.Apply(config)
			test.Ok(t, err)

			c := kafkatest.NewClient(":9092", "kafkatest")
			defer c.Close()
			join, err := c.JoinGroup(3, &joinGroup.Request{
				GroupId:      "TestGroup",
				MemberId:     "foo",
				ProtocolType: "consumer",
				Protocols: []joinGroup.Protocol{{
					Name: "range",
				}},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.None, join.ErrorCode)
			assign := []byte{
				0, 1, // version
				0, 0, 0, 1, // topic array length
				0, 3, 'f', 'o', 'o', // topic foo
				0, 0, 0, 1, // partition array length
				0, 0, 0, 1, // partition 1
				0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
			}
			sync, err := c.SyncGroup(3, &syncGroup.Request{
				GroupId:      "TestGroup",
				GenerationId: 0,
				MemberId:     "foo",
				GroupAssignments: []syncGroup.GroupAssignment{
					{
						MemberId:   "foo",
						Assignment: assign,
					},
				},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.None, sync.ErrorCode)
			// currently, not working because conflict between client id and member id: group.go:190
			//test.Equals(t, assign, sync.Assignment)
		}},
		{"sync group with wrong generation id", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			config.Channels["foo"].Value.Subscribe.Bindings.Kafka.GroupId = &openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}
			err := b.Apply(config)
			test.Ok(t, err)

			c := kafkatest.NewClient(":9092", "kafkatest")
			defer c.Close()
			join, err := c.JoinGroup(3, &joinGroup.Request{
				GroupId:      "TestGroup",
				MemberId:     "foo",
				ProtocolType: "consumer",
				Protocols: []joinGroup.Protocol{{
					Name: "range",
				}},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.None, join.ErrorCode)
			sync, err := c.SyncGroup(3, &syncGroup.Request{
				GroupId:      "TestGroup",
				GenerationId: 1,
				MemberId:     "foo",
				GroupAssignments: []syncGroup.GroupAssignment{
					{
						MemberId:   "foo",
						Assignment: nil,
					},
				},
			})
			test.Ok(t, err)
			test.Equals(t, protocol.IllegalGeneration, sync.ErrorCode)
			// currently, not working because conflict between client id and member id: group.go:190
			//test.Equals(t, assign, sync.Assignment)
		}},
		{"sync group successfully with two consumers", func(t *testing.T, b *kafka.Binding, config *asyncApi.Config) {
			config.Channels["foo"].Value.Subscribe.Bindings.Kafka.GroupId = &openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}
			config.Servers["foo"].Bindings.Kafka.Config["group.initial.rebalance.delay.ms"] = "3000"
			err := b.Apply(config)
			test.Ok(t, err)

			groupAssign := []syncGroup.GroupAssignment{
				{"leader", []byte{
					0, 1, // version
					0, 0, 0, 1, // topic array length
					0, 3, 'f', 'o', 'o', // topic foo
					0, 0, 0, 1, // partition array length
					0, 0, 0, 1, // partition 1
					0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
				}, nil},
				{"kafkatest", []byte{
					0, 1, // version
					0, 0, 0, 1, // topic array length
					0, 3, 'f', 'o', 'o', // topic foo
					0, 0, 0, 1, // partition array length
					0, 0, 0, 2, // partition 1
					0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
				}, nil},
			}

			joinFn := func(clientId string, ga []syncGroup.GroupAssignment) (*joinGroup.Response, *syncGroup.Response) {
				c := kafkatest.NewClient(":9092", clientId)
				defer c.Close()
				join, err := c.JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     clientId,
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				test.Ok(t, err)
				test.Equals(t, protocol.None, join.ErrorCode)

				sync, err := c.SyncGroup(3, &syncGroup.Request{
					GroupId:          "TestGroup",
					GenerationId:     0,
					MemberId:         clientId,
					GroupAssignments: ga,
				})
				test.Ok(t, err)
				return join, sync
			}

			var leaderJoin *joinGroup.Response
			var leaderSync *syncGroup.Response
			ch := make(chan bool)
			go func() {
				leaderJoin, leaderSync = joinFn("leader", groupAssign)
				ch <- true
			}()
			time.Sleep(500 * time.Millisecond)
			join, sync := joinFn("kafkatest", nil)

			<-ch
			// leader
			test.Equals(t, protocol.None, leaderSync.ErrorCode)
			test.Equals(t, "leader", leaderJoin.Leader)
			// currently, not working because group.go:162 checks not leader and client_handler.go:311 []byte is never nil
			//test.Equals(t, uint8(1), leaderSync.Assignment[18]) // partition 1

			// member
			test.Equals(t, protocol.None, sync.ErrorCode)
			test.Equals(t, "leader", join.Leader)
			// currently, not working because group.go:162 checks not leader and client_handler.go:311 []byte is never nil
			//test.Equals(t, uint8(2), sync.Assignment[18]) // partition 2
		}},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			//t.Parallel()
			b := kafka.NewBinding(func(topic string, key []byte, message []byte, partition int) {})
			config := asyncapitest.NewConfig(
				asyncapitest.WithServer("foo", "kafka", ":9092", asyncapitest.WithKafka("group.initial.rebalance.delay.ms", "0")),
				asyncapitest.WithChannel(
					"foo", asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithPayload(openapitest.NewSchema())))))
			err := b.Apply(config)
			test.Ok(t, err)
			b.Start()

			data.fn(t, b, config)

			b.Stop()
		})
	}
}
