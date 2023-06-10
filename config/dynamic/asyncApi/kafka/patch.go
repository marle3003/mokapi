package kafka

func (b BrokerBindings) Patch(patch BrokerBindings) {
	for k, v := range patch.Config {
		if c, ok := b.Config[k]; !ok || len(c) == 0 {
			b.Config[k] = v
		}
	}
}

func (t *TopicBindings) Patch(patch TopicBindings) {
	if t.Partitions == 0 {
		t.Partitions = patch.Partitions
	}
	if t.RetentionBytes == 0 {
		t.RetentionBytes = patch.RetentionBytes
	}
	if t.RetentionMs == 0 {
		t.RetentionMs = patch.RetentionMs
	}
	if t.SegmentBytes == 0 {
		t.SegmentBytes = patch.SegmentBytes
	}
	if t.SegmentMs == 0 {
		t.SegmentMs = patch.SegmentMs
	}
}

func (m *MessageBinding) Patch(patch MessageBinding) {
	if m.Key == nil {
		m.Key = patch.Key
	} else {
		m.Key.Patch(patch.Key)
	}
}
