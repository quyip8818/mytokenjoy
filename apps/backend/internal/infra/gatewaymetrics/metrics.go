// Package gatewaymetrics provides lightweight gateway observability counters.
// Follows the same pattern as infra/ingestmetrics: atomic counters + Snapshot().
package gatewaymetrics

import (
	"sync/atomic"
	"time"
)

// Snapshot is the JSON-serializable metrics snapshot for the gateway.
type Snapshot struct {
	RequestsAllowed     int64 `json:"gateway_requests_allowed_total"`
	RequestsRejected    int64 `json:"gateway_requests_rejected_total"`
	RequestsRateLimited int64 `json:"gateway_requests_rate_limited_total"`
	PrecheckAvgUs       int64 `json:"gateway_precheck_avg_us"`
	PrecheckMaxUs       int64 `json:"gateway_precheck_max_us"`
	PrecheckCount       int64 `json:"gateway_precheck_count"`
}

// Recorder records gateway metrics.
type Recorder interface {
	RecordAllowed()
	RecordRejected()
	RecordRateLimited()
	RecordPrecheckDuration(d time.Duration)
	Snapshot() Snapshot
}

type recorder struct {
	allowed     atomic.Int64
	rejected    atomic.Int64
	rateLimited atomic.Int64
	precheckSum atomic.Int64 // microseconds total
	precheckMax atomic.Int64 // microseconds
	precheckN   atomic.Int64
}

func NewRecorder() Recorder { return &recorder{} }

func (r *recorder) RecordAllowed()     { r.allowed.Add(1) }
func (r *recorder) RecordRejected()    { r.rejected.Add(1) }
func (r *recorder) RecordRateLimited() { r.rateLimited.Add(1) }

func (r *recorder) RecordPrecheckDuration(d time.Duration) {
	us := d.Microseconds()
	r.precheckSum.Add(us)
	r.precheckN.Add(1)
	for {
		cur := r.precheckMax.Load()
		if us <= cur || r.precheckMax.CompareAndSwap(cur, us) {
			break
		}
	}
}

func (r *recorder) Snapshot() Snapshot {
	n := r.precheckN.Load()
	var avg int64
	if n > 0 {
		avg = r.precheckSum.Load() / n
	}
	return Snapshot{
		RequestsAllowed:     r.allowed.Load(),
		RequestsRejected:    r.rejected.Load(),
		RequestsRateLimited: r.rateLimited.Load(),
		PrecheckAvgUs:       avg,
		PrecheckMaxUs:       r.precheckMax.Load(),
		PrecheckCount:       n,
	}
}

type noop struct{}

func (noop) RecordAllowed()                       {}
func (noop) RecordRejected()                      {}
func (noop) RecordRateLimited()                   {}
func (noop) RecordPrecheckDuration(time.Duration) {}
func (noop) Snapshot() Snapshot                   { return Snapshot{} }

// Noop is a no-op recorder for tests or disabled gateway.
func Noop() Recorder { return noop{} }
