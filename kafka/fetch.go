package kafka

import (
	"math"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/fetch"
	"time"
)

type fetchData struct {
	fetchOffset int64
	batch       protocol.RecordBatch
	maxBytes    int
	error       protocol.ErrorCode
	offset      int64
	startOffset int64
}

func (b *Broker) fetch(rw protocol.ResponseWriter, req *protocol.Request) error {
	f := req.Message.(*fetch.Request)
	start := time.Now().Add(time.Duration(f.MaxWaitMs-200) * time.Millisecond) // -200: work load time

	topics := make(map[string]map[int32]*fetchData)
	size := int32(0)
	for {
		for _, rt := range f.Topics {
			t, ok := topics[rt.Name]
			if !ok {
				t = make(map[int32]*fetchData)
				topics[rt.Name] = t
			}
			topic := b.Store.Topic(rt.Name)
			for _, rp := range rt.Partitions {
				data, ok := t[rp.Index]
				if !ok {
					data = &fetchData{
						fetchOffset: rp.FetchOffset,
						batch:       protocol.NewRecordBatch(),
						maxBytes:    int(rp.MaxBytes),
					}
					t[rp.Index] = data
				} else if data.error != protocol.None {
					continue
				}

				if topic == nil {
					data.error = protocol.UnknownTopicOrPartition
					continue
				}

				p := topic.Partition(int(rp.Index))
				if p == nil {
					data.error = protocol.UnknownTopicOrPartition
					continue
				}

				data.offset = p.Offset()
				data.startOffset = p.StartOffset()

				var batch protocol.RecordBatch
				batch, data.error = p.Read(data.fetchOffset, data.maxBytes)
				batchSize := batch.Size()
				size += int32(batchSize)
				data.maxBytes -= batchSize
				data.fetchOffset += int64(len(batch.Records))
				data.batch.Records = append(data.batch.Records, batch.Records...)
			}
		}

		if time.Now().After(start) || size >= f.MinBytes {
			break
		}

		sleep := math.Floor(0.2 * float64(f.MaxWaitMs))
		sleep = math.Min(sleep, 100)
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}

	res := &fetch.Response{Topics: make([]fetch.ResponseTopic, 0)}
	for name, topic := range topics {
		resTopic := fetch.ResponseTopic{Name: name, Partitions: make([]fetch.ResponsePartition, 0, len(topic))}
		for index, data := range topic {
			resPar := fetch.ResponsePartition{
				Index:                index,
				HighWatermark:        data.offset,
				LastStableOffset:     data.offset,
				LogStartOffset:       data.startOffset,
				PreferredReadReplica: -1,
				RecordSet:            data.batch,
				ErrorCode:            data.error,
			}
			resTopic.Partitions = append(resTopic.Partitions, resPar)
		}
		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
