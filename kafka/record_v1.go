package kafka

import (
	"fmt"
	"mokapi/buffer"
)

func (rb *RecordBatch) readFromV1(d *Decoder) error {
	pb := buffer.NewPageBuffer()
	defer pb.Unref()

	for d.leftSize > 0 {
		r := Record{}
		r.Offset = d.ReadInt64()
		size := d.ReadInt32()
		crc := d.ReadInt32()
		magic := d.ReadInt8()
		attributes := Attributes(d.ReadInt8())
		timestamp := d.ReadInt64()
		r.Time = ToTime(timestamp)

		_ = size
		_ = crc
		_ = magic

		if attributes.Compression() != 0 {
			return fmt.Errorf("compression currently not supported")
		}

		keyOffset := pb.Size()
		keyLength := d.ReadInt32()
		if keyLength > 0 {
			d.writeTo(pb, int(keyLength))
			r.Key = pb.Slice(keyOffset, int(keyLength))
		}

		valueOffset := pb.Size()
		valueLength := d.ReadInt32()
		if valueLength > 0 {
			d.writeTo(pb, int(valueLength))
			r.Value = pb.Slice(valueOffset, int(valueLength))
		}

		rb.Records = append(rb.Records, &r)
	}

	return nil
}
