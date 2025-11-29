package kafka

import (
	"fmt"
	"mokapi/buffer"
)

func (rb *RecordBatch) readFromV2(d *Decoder) error {
	// partition base offset of following records
	baseOffset := d.ReadInt64()
	d.ReadInt32() // message size
	d.ReadInt32() // leader epoch
	d.ReadInt8()  // magic byte version
	d.ReadInt32() // checksum
	attributes := Attributes(d.ReadInt16())
	d.ReadInt32() // lastOffsetDelta
	firstTimestamp := d.ReadInt64()
	d.ReadInt64()                  // maxTimestamp
	producerId := d.ReadInt64()    // producer ID
	producerEpoch := d.ReadInt16() // producer epoch
	sequence := d.ReadInt32()      // baseSequence
	numRecords := d.ReadInt32()

	if attributes.Compression() != 0 {
		return fmt.Errorf("compression currently not supported")
	}

	pb := buffer.NewPageBuffer()
	defer pb.Unref()
	rb.Records = make([]*Record, numRecords)
	for i := range rb.Records {
		r := &Record{
			ProducerId:     producerId,
			ProducerEpoch:  producerEpoch,
			SequenceNumber: sequence,
		}
		sequence++
		rb.Records[i] = r
		d.ReadVarInt() // record size
		d.ReadInt8()   // attributes

		timestampDelta := d.ReadVarInt()
		offsetDelta := d.ReadVarInt()
		r.Offset = baseOffset + offsetDelta
		r.Time = ToTime(firstTimestamp + timestampDelta)

		keyOffset := pb.Size()
		var keyLength int64
		keyLength = d.ReadVarInt()
		if keyLength > 0 {
			d.writeTo(pb, int(keyLength))
			r.Key = pb.Slice(keyOffset, keyOffset+int(keyLength))
		}

		valueOffset := pb.Size()
		valueLength := d.ReadVarInt()
		if valueLength > 0 {
			d.writeTo(pb, int(valueLength))
			r.Value = pb.Slice(valueOffset, valueOffset+int(valueLength))
		}

		headerLen := d.ReadVarInt()
		if headerLen > 0 {
			r.Headers = make([]RecordHeader, headerLen)
			for i := range r.Headers {
				r.Headers[i] = RecordHeader{
					Key:   d.ReadVarString(),
					Value: d.ReadVarBytes(),
				}
			}
		}
	}

	return nil
}
