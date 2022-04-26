package discovery

import (
	"context"
)

type Discoverer interface {
	DiscoverTargets(ctx context.Context) ([]string, error)
}
