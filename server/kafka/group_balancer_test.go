package kafka_test

import (
	"mokapi/server/kafka/kafkatest"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/syncGroup"
	"mokapi/test"
	"testing"
	"time"
)

func TestGroupBalancing(t *testing.T) {
	t.Parallel()
	testdata := []struct {
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
					MemberId:     "foo",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name:     "range",
						MetaData: meta,
					}},
				})
				test.Ok(t, err)
				test.Equals(t, protocol.None, join.ErrorCode)
				test.Equals(t, "foo", join.Leader)
				test.Equals(t, "range", join.ProtocolName)
				// currently, not correct because conflict between client id and member id
				test.Equals(t, "foo", join.MemberId)
				test.Equals(t, "foo", join.Members[0].MemberId)
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
					c := kafkatest.NewClient(b.Listener.Addr().String(), "kafkatest")
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
				test.Equals(t, protocol.None, member.ErrorCode)
				test.Equals(t, "foo1", member.Leader)
				test.Equals(t, "range", member.ProtocolName)
				// currently, not correct because conflict between client id and member id
				test.Equals(t, "foo2", member.MemberId)
				test.Equals(t, 0, len(member.Members))

				leader := <-ch
				test.Ok(t, err)
				test.Equals(t, protocol.None, leader.ErrorCode)
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
				test.Equals(t, protocol.IllegalGeneration, sync.ErrorCode)
			}},
		{"sync group but joining state",
			func(t *testing.T, b *kafkatest.Broker) {
				ch := make(chan *joinGroup.Response)
				go func() {
					c := kafkatest.NewClient(b.Listener.Addr().String(), "kafkatest")
					defer c.Close()
					join, err := c.JoinGroup(3, &joinGroup.Request{
						GroupId:      "TestGroup",
						MemberId:     "foo",
						ProtocolType: "consumer",
						Protocols: []joinGroup.Protocol{{
							Name: "range",
						}},
					})
					if err != nil {
						panic(err)
					}
					ch <- join
				}()

				time.Sleep(500 * time.Millisecond)
				sync, err := b.Client().SyncGroup(3, &syncGroup.Request{
					GroupId:      "TestGroup",
					GenerationId: 0,
					MemberId:     "foo2",
				})
				test.Ok(t, err)
				test.Equals(t, protocol.RebalanceInProgress, sync.ErrorCode)
				// wait for join response
				j := <-ch
				test.Equals(t, protocol.None, j.ErrorCode)
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
				test.Equals(t, protocol.None, join.ErrorCode)
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
				test.Equals(t, protocol.None, sync.ErrorCode)
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
				test.Equals(t, protocol.None, join.ErrorCode)
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
				test.Equals(t, protocol.IllegalGeneration, sync.ErrorCode)
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

				joinFn := func(clientId string, ga []syncGroup.GroupAssignment) (*joinGroup.Response, *syncGroup.Response) {
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
						panic(err)
					}

					sync, err := c.SyncGroup(3, &syncGroup.Request{
						GroupId:          "TestGroup",
						GenerationId:     0,
						MemberId:         clientId,
						GroupAssignments: ga,
					})
					if err != nil {
						panic(err)
					}
					return join, sync
				}

				var leaderJoin *joinGroup.Response
				var leaderSync *syncGroup.Response
				ch := make(chan bool)
				go func() {
					leaderJoin, leaderSync = joinFn("leader", groupAssign)
					ch <- true
				}()
				time.Sleep(1000 * time.Millisecond)
				join, sync := joinFn("member", nil)

				<-ch
				// leader
				test.Equals(t, protocol.None, leaderSync.ErrorCode)
				test.Equals(t, "leader", leaderJoin.Leader)
				test.Equals(t, "leader", leaderJoin.MemberId)
				test.Equals(t, uint8(1), leaderSync.Assignment[18]) // partition 1

				// member
				test.Equals(t, protocol.None, sync.ErrorCode)
				test.Equals(t, "leader", join.Leader)
				test.Equals(t, "member", join.MemberId)
				test.Equals(t, uint8(2), sync.Assignment[18]) // partition 2
			}},
	}

	for _, data := range testdata {
		d := data
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			b := kafkatest.NewBroker()
			defer b.Close()

			d.fn(t, b)
		})
	}
}
