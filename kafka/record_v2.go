package kafka

import "fmt"

func (rb *RecordBatch) readFromV2(d *Decoder) error {
	// partition base offset of following records
	baseOffset := d.ReadInt64()
	d.ReadInt32()     // batchLength
	d.ReadInt32()     // leader epoch
	m := d.ReadInt8() // magic
	_ = m
	crc := d.ReadInt32() // checksum
	attributes := Attributes(d.ReadInt16())
	d.ReadInt32() // lastOffsetDelta
	firstTimestamp := d.ReadInt64()
	d.ReadInt64() // maxTimestamp
	producerId := d.ReadInt64()
	producerEpoch := d.ReadInt16()
	d.ReadInt32() // baseSequence
	numRecords := d.ReadInt32()

	_ = crc
	_ = producerId
	_ = producerEpoch

	if attributes.Compression() != 0 {
		return fmt.Errorf("compression currently not supported")
	}

	pb := newPageBuffer()
	defer pb.unref()
	rb.Records = make([]*Record, numRecords)
	for i := range rb.Records {
		r := &Record{}
		rb.Records[i] = r
		l := d.ReadVarInt() // length
		_ = l
		d.ReadInt8() // attributes

		timestampDelta := d.ReadVarInt()
		offsetDelta := d.ReadVarInt()
		r.Offset = baseOffset + offsetDelta
		r.Time = ToTime(firstTimestamp + timestampDelta)

		keyOffset := pb.Size()
		keyLength := d.ReadVarInt()
		if keyLength > 0 {
			d.writeTo(pb, int(keyLength))
			r.Key = pb.fragment(keyOffset, keyOffset+int(keyLength))
		}

		valueOffset := pb.Size()
		valueLength := d.ReadVarInt()
		if valueLength > 0 {
			d.writeTo(pb, int(valueLength))
			r.Value = pb.fragment(valueOffset, valueOffset+int(valueLength))
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
