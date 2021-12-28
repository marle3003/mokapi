package store

type Topic struct {
	name       string
	partitions map[int]*Partition

	validator *validator
}

func (t *Topic) Name() string {
	return t.name
}

func (t *Topic) Partition(index int) *Partition {
	if p, ok := t.partitions[index]; ok {
		return p
	}
	return nil
}

func (t *Topic) Partitions() []*Partition {
	partitions := make([]*Partition, 0, len(t.partitions))
	for _, p := range t.partitions {
		partitions = append(partitions, p)
	}
	return partitions
}
