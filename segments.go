package socket_oob

// Segments represents a series of network packets in one byte array.
// all segments have MaxSegmentSize except the last one might be smaller.
type Segments struct {
	Buf            []byte
	MaxSegmentSize int
}

func (s *Segments) Iterator() Iterator[[]byte] {
	return &segmentIterator{
		buf:            s.Buf,
		maxSegmentSize: s.MaxSegmentSize,
		index:          0,
	}
}

func (s *Segments) Split(left func(segment []byte) bool) (Segments, Segments) {
	splitIndex := len(s.Buf)
	iter := s.Iterator().(*segmentIterator)
	for iter.HasNext() {
		segment := iter.Next()
		if !left(segment) {
			splitIndex = iter.index
			break
		}
	}
	return Segments{
			Buf:            s.Buf[:splitIndex],
			MaxSegmentSize: s.MaxSegmentSize,
		}, Segments{
			Buf:            s.Buf[splitIndex:],
			MaxSegmentSize: s.MaxSegmentSize,
		}
}

func (s *Segments) Len() int {
	return len(s.Buf)
}

func (s *Segments) First() []byte {
	end := s.MaxSegmentSize
	if end > len(s.Buf) {
		end = len(s.Buf)
	}
	return s.Buf[:end]
}
