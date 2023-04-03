package middleware

import (
	"net/http"
)

type Middleware interface {
	http.Handler
	Init(http.Handler) Middleware
}
