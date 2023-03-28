package chaincontext

import "github.com/ferama/crauti/pkg/conf"

type contextKey string

const ChainContextKey contextKey = "chain-context"

// The chain context holds all the mountPoints related config
// It is easily accessed from all the middleware without requiring
// any custom variable passing and stuff
type ChainContext struct {
	Conf conf.MountPoint
}
