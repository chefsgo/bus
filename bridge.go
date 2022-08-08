package bus

import (
	"time"

	"github.com/chefsgo/chef"
)

func init() {
	chef.Register(bridge)
}

var (
	bridge = &Bridge{}
)

type (
	Bridge struct{}
)

func (this *Bridge) Request(meta chef.Metadata, timeout time.Duration) (*chef.Echo, error) {
	return module.Request(meta, timeout)
}
