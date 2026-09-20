package images

import (
	"bytes"
	"sync"
)

type Binding struct{ Owner, Installation, Asset, SourceDigest string }
type Limits struct{ LongEdge, MaxPixels, MaxDimension, MaxSourceBytes, MaxRequestBytes int }

func DefaultLimits() Limits { return Limits{2048, 40000000, 40000, 20 << 20, 10 << 20} }

type Metadata struct {
	Binding              Binding
	Width, Height        int
	MIME, Digest, Policy string
}
type preparedState struct {
	mu       sync.RWMutex
	data     []byte
	metadata Metadata
}
type Prepared struct{ state *preparedState }

func (p *Prepared) Bytes() ([]byte, error) {
	if p == nil || p.state == nil {
		return nil, ErrInvalid
	}
	p.state.mu.RLock()
	defer p.state.mu.RUnlock()
	if len(p.state.data) == 0 {
		return nil, ErrInvalid
	}
	return bytes.Clone(p.state.data), nil
}
func (p *Prepared) Info() (Metadata, bool) {
	if p == nil || p.state == nil {
		return Metadata{}, false
	}
	p.state.mu.RLock()
	defer p.state.mu.RUnlock()
	if len(p.state.data) == 0 {
		return Metadata{}, false
	}
	return p.state.metadata, true
}
func (p *Prepared) Release() {
	if p == nil || p.state == nil {
		return
	}
	p.state.mu.Lock()
	defer p.state.mu.Unlock()
	clear(p.state.data)
	p.state.data = nil
	p.state.metadata = Metadata{}
}
