package middleware

import (
	"net/http"

	"github.com/ferama/crauti/pkg/chaincontext"
)

type Middleware struct{}

func (m *Middleware) GetContext(r *http.Request) chaincontext.ChainContext {
	return chaincontext.GetChainContext(r)
}
