package store

import (
	"bufio"
	"bytes"
	"mokapi/kafka"
	"mokapi/kafka/syncGroup"
)

func (s *Store) syncgroup(rw kafka.ResponseWriter, req *kafka.Request) error {
	r := req.Message.(*syncGroup.Request)
	ctx := kafka.ClientFromContext(req.Context)

	if len(r.MemberId) == 0 {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.MemberIdRequired})
	}

	b := s.getBrokerByHost(req.Host)
	if b == nil {
		res := &syncGroup.Response{
			ErrorCode: kafka.UnknownServerError,
		}
		return rw.Write(res)
	}

	g := s.GetOrCreateGroup(r.GroupId, b.Id)

	if g.State == PreparingRebalance {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.RebalanceInProgress})
	}

	if g.Generation == nil || g.Generation.Id != int(r.GenerationId) {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.IllegalGeneration})
	}

	if _, ok := ctx.Member[r.GroupId]; !ok {
		return rw.Write(&syncGroup.Response{ErrorCode: kafka.RebalanceInProgress})
	}

	data := syncdata{
		client:       ctx,
		writer:       rw,
		protocolType: r.ProtocolType,
		protocolName: r.ProtocolName,
		generationId: r.GenerationId,
		log:          s.logRequest(req.Header),
	}

	if len(r.GroupAssignments) > 0 {
		data.assigns = make(map[string]*groupAssignment)
		for _, assign := range r.GroupAssignments {
			data.assigns[assign.MemberId] = newGroupAssignment(assign.Assignment)
		}
	}

	// balancer writes the response
	g.balancer.sync <- data

	return nil
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
