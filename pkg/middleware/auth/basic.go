package auth

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
)

type BasicAuthMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *BasicAuthMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next

	return m
}

func (m *BasicAuthMiddleware) unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}

func (m *BasicAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := chaincontext.GetChainContext(r)
	if !ctx.Conf.Middlewares.BasicAuth.IsEnabled() {
		m.next.ServeHTTP(w, r)
		return
	}

	realm := ctx.Conf.Middlewares.BasicAuth.Realm

	username, password, ok := r.BasicAuth()
	if !ok {
		m.unauthorized(w, realm)
		return
	}

	credentials := ctx.Conf.Middlewares.BasicAuth.Credentials
	passwordBytes := []byte(password)

	for _, v := range credentials {
		if v.Username == username {
			validPassword := v.Password
			validPasswordBytes := []byte(v.Password)

			// take the same amount of time if the lengths are different
			// this is required since ConstantTimeCompare returns immediately when slices
			// of different length are compared and prevents timing attack
			if len(password) != len(validPassword) {
				subtle.ConstantTimeCompare(validPasswordBytes, validPasswordBytes)
			} else {
				if subtle.ConstantTimeCompare(passwordBytes, validPasswordBytes) == 1 {
					m.next.ServeHTTP(w, r)
					return
				}
			}
		}
	}

	m.unauthorized(w, realm)
}
