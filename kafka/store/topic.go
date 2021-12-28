package store

type Topic struct {
	name       string
	partitions []*Partition

	validator *validator
}

func (t *Topic) Name() string {
	return t.name
}

func (t *Topic) Partition(index int) *Partition {
	if index >= len(t.partitions) {
		return nil
	}
	return t.partitions[index]
}

func (t *Topic) Partitions() []*Partition {
	partitions := make([]*Partition, 0, len(t.partitions))
	for _, p := range t.partitions {
		partitions = append(partitions, p)
	}
	return partitions
}

func (t *Topic) delete() {
	for _, p := range t.partitions {
		p.delete()
	}
}
