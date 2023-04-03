package utils

import (
	"net/http"

	"github.com/ferama/crauti/pkg/middleware/collector"
)

// Usefull if you should break a chain prematurely but
// you still need to emit available logs
// It should be called like:
//
//	utils.EmitAndReturn(w, r)
//	return
func EmitAndReturn(w http.ResponseWriter, r *http.Request) {
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	var chain http.Handler
	chain = root

	chain = (&collector.EmitterMiddleware{}).Init(chain)

	chain.ServeHTTP(w, r)
}
