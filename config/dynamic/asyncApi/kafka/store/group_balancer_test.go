package store_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/syncGroup"
	"mokapi/test"
	"testing"
	"time"
)

func TestGroupBalancing(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, b *kafkatest.Broker)
	}{
		{"join group",
			func(t *testing.T, b *kafkatest.Broker) {
				meta := []byte{
					0, 1, // version
					0, 0, 0, 1, // topic array length
					0, 3, 'f', 'o', 'o', // topic foo
					0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
				}
				join, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name:     "range",
						MetaData: meta,
					}},
				})
				test.Ok(t, err)
				test.Equals(t, kafka.None, join.ErrorCode)
				test.Equals(t, join.MemberId, join.Leader)
				test.Equals(t, "range", join.ProtocolName)
				// currently, not correct because conflict between client id and member id
				test.Assert(t, len(join.MemberId) > 0, "no member id assigned")
				test.Equals(t, join.MemberId, join.Members[0].MemberId)
				test.Equals(t, meta, join.Members[0].MetaData)
			}},
		{"two members join same group",
			func(t *testing.T, b *kafkatest.Broker) {
				meta := []byte{
					0, 1, // version
					0, 0, 0, 1, // topic array length
					0, 3, 'f', 'o', 'o', // topic foo
					0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
				}
				ch := make(chan *joinGroup.Response, 1)
				go func() {
					c := kafkatest.NewClient(b.Addr, "kafkatest")
					defer c.Close()
					join, err := c.JoinGroup(3, &joinGroup.Request{
						GroupId:      "TestGroup",
						MemberId:     "foo1",
						ProtocolType: "consumer",
						Protocols: []joinGroup.Protocol{{
							Name:     "range",
							MetaData: meta,
						}},
					})
					if err != nil {
						panic(err)
					}
					ch <- join
				}()
				time.Sleep(500 * time.Millisecond)
				member, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     "foo2",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name:     "range",
						MetaData: meta,
					}},
				})

				test.Ok(t, err)
				test.Equals(t, kafka.None, member.ErrorCode)
				test.Equals(t, "foo1", member.Leader)
				test.Equals(t, "range", member.ProtocolName)
				// currently, not correct because conflict between client id and member id
				test.Equals(t, "foo2", member.MemberId)
				test.Equals(t, 0, len(member.Members))

				leader := <-ch
				test.Ok(t, err)
				test.Equals(t, kafka.None, leader.ErrorCode)
				test.Equals(t, "foo1", leader.Leader)
				test.Equals(t, "range", leader.ProtocolName)
				// currently, not correct because conflict between client id and member id
				test.Equals(t, "foo1", leader.MemberId)
				test.Equals(t, 2, len(leader.Members))
				test.Equals(t, "foo1", leader.Members[0].MemberId)
				test.Equals(t, meta, leader.Members[0].MetaData)
			}},
		{"sync group but not member",
			func(t *testing.T, b *kafkatest.Broker) {
				sync, err := b.Client().SyncGroup(3, &syncGroup.Request{
					GroupId:      "TestGroup",
					GenerationId: 0,
					MemberId:     "foo",
				})
				test.Ok(t, err)
				test.Equals(t, kafka.IllegalGeneration, sync.ErrorCode)
			}},
		{"sync group but joining state",
			func(t *testing.T, b *kafkatest.Broker) {
				ch := make(chan *joinGroup.Response)
				go func() {
					c := kafkatest.NewClient(b.Addr, "kafkatest")
					defer c.Close()
					join, _ := c.JoinGroup(3, &joinGroup.Request{
						GroupId:      "TestGroup",
						MemberId:     "foo",
						ProtocolType: "consumer",
						Protocols: []joinGroup.Protocol{{
							Name: "range",
						}},
					})
					ch <- join
				}()

				time.Sleep(500 * time.Millisecond)
				sync, err := b.Client().SyncGroup(3, &syncGroup.Request{
					GroupId:      "TestGroup",
					GenerationId: 0,
					MemberId:     "foo2",
				})
				test.Ok(t, err)
				test.Equals(t, kafka.RebalanceInProgress, sync.ErrorCode)
				// wait for join response
				<-ch
			}},
		{"sync group successfully",
			func(t *testing.T, b *kafkatest.Broker) {
				join, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     "foo",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				test.Ok(t, err)
				test.Equals(t, kafka.None, join.ErrorCode)
				assign := []byte{
					0, 1, // version
					0, 0, 0, 1, // topic array length
					0, 3, 'f', 'o', 'o', // topic foo
					0, 0, 0, 1, // partition array length
					0, 0, 0, 1, // partition 1
					0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
				}
				sync, err := b.Client().SyncGroup(3, &syncGroup.Request{
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
				test.Equals(t, kafka.None, sync.ErrorCode)
				test.Equals(t, assign, sync.Assignment)
			}},
		{"sync group with wrong generation id",
			func(t *testing.T, b *kafkatest.Broker) {
				join, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     "foo",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				test.Ok(t, err)
				test.Equals(t, kafka.None, join.ErrorCode)
				sync, err := b.Client().SyncGroup(3, &syncGroup.Request{
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
				test.Equals(t, kafka.IllegalGeneration, sync.ErrorCode)
			}},
		{"sync group successfully with two consumers",
			func(t *testing.T, b *kafkatest.Broker) {
				groupAssign := []syncGroup.GroupAssignment{
					{"leader", []byte{
						0, 1, // version
						0, 0, 0, 1, // topic array length
						0, 3, 'f', 'o', 'o', // topic foo
						0, 0, 0, 1, // partition array length
						0, 0, 0, 1, // partition 1
						0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
					}, nil},
					{"member", []byte{
						0, 1, // version
						0, 0, 0, 1, // topic array length
						0, 3, 'f', 'o', 'o', // topic foo
						0, 0, 0, 1, // partition array length
						0, 0, 0, 2, // partition 1
						0, 0, 0, 3, 0x01, 0x02, 0x03, // userdata
					}, nil},
				}

				joinFn := func(clientId string, ga []syncGroup.GroupAssignment) (*joinGroup.Response, *syncGroup.Response, error) {
					c := kafkatest.NewClient(b.Client().Addr, clientId)
					defer c.Close()
					join, err := c.JoinGroup(3, &joinGroup.Request{
						GroupId:      "TestGroup",
						MemberId:     clientId,
						ProtocolType: "consumer",
						Protocols: []joinGroup.Protocol{{
							Name: "range",
						}},
					})
					if err != nil {
						return nil, nil, err
					}
					sync, err := c.SyncGroup(3, &syncGroup.Request{
						GroupId:          "TestGroup",
						GenerationId:     0,
						MemberId:         clientId,
						GroupAssignments: ga,
					})
					if err != nil {
						return nil, nil, err
					}
					return join, sync, nil
				}

				var leaderJoin *joinGroup.Response
				var leaderSync *syncGroup.Response
				var joinErr error
				ch := make(chan bool)
				go func() {
					leaderJoin, leaderSync, joinErr = joinFn("leader", groupAssign)
					ch <- true
				}()
				// ensure order of members and thus leader election. first member is leader
				time.Sleep(1 * time.Second)
				join, sync, err := joinFn("member", nil)
				test.Ok(t, err)

				<-ch
				test.Ok(t, joinErr)
				// leader
				test.Equals(t, kafka.None, leaderSync.ErrorCode)
				test.Equals(t, "leader", leaderJoin.Leader)
				test.Equals(t, "leader", leaderJoin.MemberId)
				test.Equals(t, uint8(1), leaderSync.Assignment[18]) // partition 1

				// member
				test.Equals(t, kafka.None, sync.ErrorCode)
				test.Equals(t, "leader", join.Leader)
				test.Equals(t, "member", join.MemberId)
				test.Equals(t, uint8(2), sync.Assignment[18]) // partition 2
			}},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := store.New(asyncapitest.NewConfig())
			b := kafkatest.NewBroker(kafkatest.WithHandler(s))
			defer b.Close()
			s.Update(asyncapitest.NewConfig(asyncapitest.WithServer("", "kafka", b.Addr)))

			tc.fn(t, b)
		})
	}
}
