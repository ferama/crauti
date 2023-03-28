package middleware

import (
	"net/http"

	"github.com/ferama/crauti/pkg/chaincontext"
)

type Middleware struct{}

func (m *Middleware) GetChainContext(r *http.Request) chaincontext.ChainContext {
	chainContext := r.Context().
		Value(chaincontext.ChainContextKey).(chaincontext.ChainContext)

	return chainContext
}
