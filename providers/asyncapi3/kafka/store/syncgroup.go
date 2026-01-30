package store

import (
	"bufio"
	"bytes"
	"mokapi/kafka"
	"mokapi/kafka/syncGroup"
)

func (s *Store) syncgroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*syncGroup.Request)
	res := &syncGroup.Response{}
	ctx := kafka.ClientFromContext(req.Context)

	data := syncdata{
		client:       ctx,
		writer:       rw,
		protocolType: r.ProtocolType,
		protocolName: r.ProtocolName,
		generationId: r.GenerationId,
	}

	if len(r.GroupAssignments) > 0 {
		data.assigns = make(map[string]*groupAssignment)
		for _, assign := range r.GroupAssignments {
			data.assigns[assign.MemberId] = newGroupAssignment(assign.Assignment)
		}
	}

	reqLog := newKafkaSyncGroupRequest(r, data.assigns)
	data.log = s.logRequest(req.Header, reqLog)

	if len(r.MemberId) != 0 {
		b := s.getBrokerByHost(req.Host)
		if b != nil {
			g := s.GetOrCreateGroup(r.GroupId, b.Id)

			if g.State != PreparingRebalance {
				if g.Generation == nil || g.Generation.Id != int(r.GenerationId) {
					res.ErrorCode = kafka.IllegalGeneration
				} else {
					if _, ok := ctx.Member[r.GroupId]; !ok {
						res.ErrorCode = kafka.RebalanceInProgress
					} else {
						// balancer writes the response
						g.balancer.sync <- data
						return nil
					}
				}
			} else {
				res.ErrorCode = kafka.RebalanceInProgress
			}
		} else {
			res.ErrorCode = kafka.UnknownServerError
		}
	} else {
		res.ErrorCode = kafka.MemberIdRequired
	}

	go func() {
		resLog := &KafkaSyncGroupResponse{}
		resLog.ErrorCode = res.ErrorCode.String()

		s.logRequest(req.Header, reqLog)(&KafkaRequestLogEvent{
			Response: resLog,
		})
	}()

	return rw.Write(res)
}

func newGroupAssignment(b []byte) *groupAssignment {
	g := &groupAssignment{}
	g.raw = b
	r := bufio.NewReader(bytes.NewReader(b))
	d := kafka.NewDecoder(r, len(b))
	g.version = d.ReadInt16()

	g.topics = make(map[string][]int)
	n := int(d.ReadInt32())
	for i := 0; i < n; i++ {
		key := d.ReadString()
		value := make([]int, 0)

		nPartition := int(d.ReadInt32())
		for j := 0; j < nPartition; j++ {
			index := d.ReadInt32()
			value = append(value, int(index))
		}
		g.topics[key] = value
	}

	g.userData = d.ReadBytes()

	return g
}

func newKafkaSyncGroupRequest(req *syncGroup.Request, assigns map[string]*groupAssignment) *KafkaSyncGroupRequest {
	r := &KafkaSyncGroupRequest{
		GroupName:        req.GroupId,
		MemberId:         req.MemberId,
		ProtocolType:     req.ProtocolType,
		GroupAssignments: map[string]KafkaSyncGroupAssignment{},
	}
	for m, a := range assigns {
		r.GroupAssignments[m] = KafkaSyncGroupAssignment{
			Version: a.version,
			Topics:  a.topics,
		}
	}
	return r
}
