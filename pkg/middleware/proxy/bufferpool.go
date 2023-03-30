package proxy

import (
	"sync"
)

type bufferPool struct {
	pool sync.Pool
}

func newPool() *bufferPool {
	p := &bufferPool{
		pool: sync.Pool{
			New: func() any {
				bp := make([]byte, 32*1024)
				log.Print("== pool new allocation: len ", len(bp))
				return &bp
			},
		},
	}
	return p
}

func (p *bufferPool) Get() []byte {
	buf := p.pool.Get().(*[]byte)
	// log.Print("== get from pool: len ", len(*buf))
	return *buf
}
func (p *bufferPool) Put(b []byte) {
	// log.Print("== put into pool")
	p.pool.Put(&b)
}
