package requester

import (
	"time"

	"github.com/tcnksm/go-httpstat"
)

type Request struct {
	index int
}
type Response struct {
	Stat       httpstat.Result
	Timeout    bool
	Error      bool
	StatusCode int
	Status     string
}

func (r Response) RequestLatency() time.Duration {
	return r.Stat.DNSLookup + r.Stat.TCPConnection
}

func (r Response) ServerLatency() time.Duration {
	return r.Stat.TLSHandshake + r.Stat.ServerProcessing
}

func (r Response) TotalLatency() time.Duration {
	return r.RequestLatency() + r.ServerLatency()
}
