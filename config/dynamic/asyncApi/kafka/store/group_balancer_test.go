package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/syncGroup"
	"testing"
	"time"
)

func TestGroupBalancing(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, b *kafkatest.Broker, s *store.Store)
	}{
		{
			name: "join group",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
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
				require.NoError(t, err)
				require.Equal(t, kafka.None, join.ErrorCode)
				require.Equal(t, join.MemberId, join.Leader)
				require.Equal(t, "range", join.ProtocolName)
				// currently, not correct because conflict between client id and member id
				require.True(t, len(join.MemberId) > 0, "no member id assigned")
				require.Equal(t, join.MemberId, join.Members[0].MemberId)
				require.Equal(t, meta, join.Members[0].MetaData)
			}},
		{
			name: "join without memberId",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
				join, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				require.NoError(t, err)
				require.Equal(t, kafka.None, join.ErrorCode)
			},
		},
		{
			name: "two consumer join without memberId",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
				ch := make(chan *joinGroup.Response, 2)
				runJoin := func() {
					c := kafkatest.NewClient(b.Addr, "kafkatest")
					defer c.Close()
					join, err := c.JoinGroup(3, &joinGroup.Request{
						GroupId:      "TestGroup",
						ProtocolType: "consumer",
						Protocols: []joinGroup.Protocol{{
							Name: "range",
						}},
					})
					if err != nil {
						panic(err)
					}
					ch <- join
				}

				go runJoin()
				go runJoin()

				j1 := <-ch
				j2 := <-ch

				require.Equal(t, kafka.None, j1.ErrorCode)
				require.Equal(t, kafka.None, j2.ErrorCode)

				g, _ := s.Group("TestGroup")
				require.Len(t, g.Generation.Members, 2)
			},
		},
		{
			name: "two members join same group",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
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

				require.NoError(t, err)
				require.Equal(t, kafka.None, member.ErrorCode)
				require.Equal(t, "foo1", member.Leader)
				require.Equal(t, "range", member.ProtocolName)
				// currently, not correct because conflict between client id and member id
				require.Equal(t, "foo2", member.MemberId)
				require.Equal(t, 0, len(member.Members))

				leader := <-ch
				require.NoError(t, err)
				require.Equal(t, kafka.None, leader.ErrorCode)
				require.Equal(t, "foo1", leader.Leader)
				require.Equal(t, "range", leader.ProtocolName)
				// currently, not correct because conflict between client id and member id
				require.Equal(t, "foo1", leader.MemberId)
				require.Equal(t, 2, len(leader.Members))
				require.Equal(t, "foo1", leader.Members[0].MemberId)
				require.Equal(t, meta, leader.Members[0].MetaData)
			}},
		{
			name: "sync group but not member",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
				sync, err := b.Client().SyncGroup(3, &syncGroup.Request{
					GroupId:      "TestGroup",
					GenerationId: 0,
					MemberId:     "foo",
				})
				require.NoError(t, err)
				require.Equal(t, kafka.IllegalGeneration, sync.ErrorCode)
			}},
		{
			name: "sync group but joining state",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
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
				require.NoError(t, err)
				require.Equal(t, kafka.RebalanceInProgress, sync.ErrorCode)
				// wait for join response
				<-ch
			}},
		{
			name: "sync group successfully",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
				join, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     "foo",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				require.NoError(t, err)
				require.Equal(t, kafka.None, join.ErrorCode)
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
				require.NoError(t, err)
				require.Equal(t, kafka.None, sync.ErrorCode)
				require.Equal(t, assign, sync.Assignment)

				g, ok := s.Group("TestGroup")
				require.True(t, ok, "group must exist")
				require.NotNil(t, g.Generation, "generation must exist")
				require.Len(t, g.Generation.Members, 1, "group must contain one members")
				require.Contains(t, g.Generation.Members, "foo")
				m := g.Generation.Members["foo"]
				require.Len(t, m.Partitions, 1)
				require.Contains(t, m.Partitions, "foo")
				require.Equal(t, m.Partitions["foo"], []int{1})
			}},
		{
			name: "sync group with wrong generation id",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
				join, err := b.Client().JoinGroup(3, &joinGroup.Request{
					GroupId:      "TestGroup",
					MemberId:     "foo",
					ProtocolType: "consumer",
					Protocols: []joinGroup.Protocol{{
						Name: "range",
					}},
				})
				require.NoError(t, err)
				require.Equal(t, kafka.None, join.ErrorCode)
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
				require.NoError(t, err)
				require.Equal(t, kafka.IllegalGeneration, sync.ErrorCode)
			}},
		{
			name: "sync group successfully with two consumers",
			test: func(t *testing.T, b *kafkatest.Broker, s *store.Store) {
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
						0, 0, 0, 2, // partition 2
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
				require.NoError(t, err)

				<-ch
				require.NoError(t, joinErr)
				// leader
				require.Equal(t, kafka.None, leaderSync.ErrorCode)
				require.Equal(t, "leader", leaderJoin.Leader)
				require.Equal(t, "leader", leaderJoin.MemberId)
				require.Equal(t, uint8(1), leaderSync.Assignment[18]) // partition 1

				// member
				require.Equal(t, kafka.None, sync.ErrorCode)
				require.Equal(t, "leader", join.Leader)
				require.Equal(t, "member", join.MemberId)
				require.Equal(t, uint8(2), sync.Assignment[18]) // partition 2

				// group
				g, ok := s.Group("TestGroup")
				require.True(t, ok, "group must exist")
				require.NotNil(t, g.Generation, "generation must exist")
				require.Len(t, g.Generation.Members, 2, "group must contain two members")
				require.Contains(t, g.Generation.Members, "leader")
				require.Contains(t, g.Generation.Members, "member")

				// leader
				m := g.Generation.Members["leader"]
				require.Len(t, m.Partitions, 1)
				require.Contains(t, m.Partitions, "foo")
				require.Equal(t, m.Partitions["foo"], []int{1})

				// member
				m = g.Generation.Members["member"]
				require.Len(t, m.Partitions, 1)
				require.Contains(t, m.Partitions, "foo")
				require.Equal(t, m.Partitions["foo"], []int{2})
			}},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := store.New(asyncapitest.NewConfig(), enginetest.NewEngine())
			defer s.Close()
			b := kafkatest.NewBroker(kafkatest.WithHandler(s))
			defer b.Close()
			s.Update(asyncapitest.NewConfig(asyncapitest.WithServer("", "kafka", b.Addr)))

			tc.test(t, b, s)
		})
	}
}
