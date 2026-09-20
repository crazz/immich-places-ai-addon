package images

import (
	"bytes"
	"errors"
)

var ErrLimit = errors.New("image preparation limit exceeded")

func (l Limits) valid() bool {
	max := DefaultLimits()
	return l.LongEdge > 0 && l.LongEdge <= max.LongEdge &&
		l.MaxPixels > 0 && l.MaxPixels <= max.MaxPixels &&
		l.MaxDimension > 0 && l.MaxDimension <= max.MaxDimension &&
		l.MaxSourceBytes > 0 && l.MaxSourceBytes <= max.MaxSourceBytes &&
		l.MaxRequestBytes > 0 && l.MaxRequestBytes <= max.MaxRequestBytes
}

type boundedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(data []byte) (int, error) {
	if len(data) > b.limit-b.Len() {
		return 0, ErrLimit
	}
	return b.Buffer.Write(data)
}
