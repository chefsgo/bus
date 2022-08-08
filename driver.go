package bus

import "time"

type (
	// Driver 数据驱动
	Driver interface {
		Connect(name string, config Config) (Connect, error)
	}
	Health struct {
		Workload int64
	}

	Delegate interface {
		Serve(name string, data []byte, callback Callback)
	}

	// Connect 连接
	Connect interface {
		Open() error
		Health() (Health, error)
		Close() error

		Accept(Delegate) error
		Register(name string) error

		Start() error

		Request(name string, data []byte, timeout time.Duration) ([]byte, error)
	}
	Callback = func([]byte, error)
)
