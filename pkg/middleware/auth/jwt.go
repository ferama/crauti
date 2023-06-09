package auth

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

// Checks for JWT token validity and puts claims into context
// Uses jwks standard URL
// Example using keyclaok:
//
//	https://keycloak.url/realms/test/protocol/openid-connect/certs
type JWTAuthMiddleware struct {
	middleware.Middleware

	next http.Handler

	jwks map[string]*keyfunc.JWKS
	mu   sync.Mutex
}

func (m *JWTAuthMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next

	m.jwks = make(map[string]*keyfunc.JWKS)
	return m
}

func (m *JWTAuthMiddleware) getJWKS(jwksURL string) *keyfunc.JWKS {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item, ok := m.jwks[jwksURL]; ok {
		return item
	}
	jwks, _ := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			log.Err(err).Send()
		},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	})

	m.jwks[jwksURL] = jwks

	return jwks
}

func (m *JWTAuthMiddleware) serverErrorResponse(val string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s\n", val)
}

func (m *JWTAuthMiddleware) unauthorizedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "unauthorized\n")
}

func (m *JWTAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := chaincontext.GetChainContext(r)

	// ignore http options method
	if r.Method == http.MethodOptions || ctx.Conf.Middlewares.JwksURL == "" {
		m.next.ServeHTTP(w, r)
		return
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		m.unauthorizedResponse(w)
		return
	}

	// extract bearer
	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		m.unauthorizedResponse(w)
		return
	}
	bearer := parts[1]

	jwks := m.getJWKS(ctx.Conf.Middlewares.JwksURL)

	// Parse the JWT.
	token, err := jwt.Parse(bearer, jwks.Keyfunc)
	if err != nil {
		m.serverErrorResponse(fmt.Sprintf("failed to parse token: %s", err), w)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ctx.Auth.Authorized = true
		ctx.Auth.JwtClaims = claims
	} else {
		m.unauthorizedResponse(w)
		return
	}

	m.next.ServeHTTP(w, r)
}
