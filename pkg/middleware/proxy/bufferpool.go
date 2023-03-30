package proxy

import (
	"net/http/httputil"
	"sync"
)

type bufferPool struct {
	httputil.BufferPool

	pool sync.Pool
}

func newPool() *bufferPool {
	p := &bufferPool{
		pool: sync.Pool{
			New: func() any {
				log.Print("== pool new allocation")
				bp := make([]byte, 0, 32*1024)
				return &bp
			},
		},
	}
	return p
}

func (p *bufferPool) Get() []byte {
	// log.Print("== get from pool")
	buf := p.pool.Get().(*[]byte)
	return *buf
}
func (p *bufferPool) Put(b []byte) {
	// log.Print("== put into pool")
	p.pool.Put(&b)
}
