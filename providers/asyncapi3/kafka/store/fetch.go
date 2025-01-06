package store

import (
	log "github.com/sirupsen/logrus"
	"math"
	"mokapi/kafka"
	"mokapi/kafka/fetch"
	"time"
)

type fetchData struct {
	fetchOffset int64
	batch       kafka.RecordBatch
	maxBytes    int
	error       kafka.ErrorCode
	offset      int64
	startOffset int64
}

func (s *Store) fetch(rw kafka.ResponseWriter, req *kafka.Request) error {
	f := req.Message.(*fetch.Request)
	start := time.Now().Add(time.Duration(f.MaxWaitMs-200) * time.Millisecond) // -200: work load time

	topics := make(map[string]map[int32]*fetchData)
	size := int32(0)
	maxSize := f.MaxBytes
	for {
		for _, rt := range f.Topics {
			t, ok := topics[rt.Name]
			if !ok {
				t = make(map[int32]*fetchData)
				topics[rt.Name] = t
			}
			topic := s.Topic(rt.Name)
			for _, rp := range rt.Partitions {
				data, ok := t[rp.Index]
				if !ok {
					data = &fetchData{
						fetchOffset: rp.FetchOffset,
						batch:       kafka.NewRecordBatch(),
						maxBytes:    int(rp.MaxBytes),
					}
					t[rp.Index] = data
				} else if data.error != kafka.None {
					continue
				}

				if topic == nil {
					log.Errorf("kafka Fetch: unknown topic %v", rt.Name)
					data.error = kafka.UnknownTopicOrPartition
					continue
				}

				p := topic.Partition(int(rp.Index))
				if p == nil {
					log.Errorf("kafka Fetch: unknown partition %v", rp.Index)
					data.error = kafka.UnknownTopicOrPartition
					continue
				}

				data.offset = p.Offset()
				data.startOffset = p.StartOffset()

				var batch kafka.RecordBatch
				batch, data.error = p.Read(data.fetchOffset, data.maxBytes)
				batchSize := batch.Size()
				size += int32(batchSize)
				if size > maxSize {
					break
				}
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
			if len(data.batch.Records) > 0 {
				resPar.RecordSet = data.batch
			}
			resTopic.Partitions = append(resTopic.Partitions, resPar)
		}
		res.Topics = append(res.Topics, resTopic)
	}

	return rw.Write(res)
}
