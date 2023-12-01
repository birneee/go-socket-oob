package socket_oob

type segmentIterator struct {
	buf            []byte
	maxSegmentSize int
	index          int
}

func (i *segmentIterator) HasNext() bool {
	return i.index < len(i.buf)
}

func (i *segmentIterator) Next() []byte {
	start := i.index
	end := start + i.maxSegmentSize
	if end > len(i.buf) {
		end = len(i.buf)
	}
	i.index += i.maxSegmentSize
	return i.buf[start:end]
}
