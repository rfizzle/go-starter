package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// poolStatter is the subset of *pgxpool.Pool the collector needs. Depending on
// the interface rather than the concrete pool keeps the collector unit-testable
// with a fake and documents that it only ever reads a stats snapshot.
type poolStatter interface {
	Stat() *pgxpool.Stat
}

// PoolCollector is a prometheus.Collector that exposes pgxpool connection-pool
// statistics. It reads pool.Stat() lazily on every scrape rather than sampling
// on a timer, so each /metrics scrape reflects the pool's live state with no
// background goroutine to manage or leak.
//
// Surfacing acquired/idle/total counts alongside the empty-acquire and
// canceled-acquire counters makes connection-pool saturation visible before it
// manifests as query timeouts: a rising empty-acquire count with acquired
// pinned at max is exhaustion.
type PoolCollector struct {
	pool poolStatter

	acquiredConns           *prometheus.Desc
	idleConns               *prometheus.Desc
	totalConns              *prometheus.Desc
	maxConns                *prometheus.Desc
	constructingConns       *prometheus.Desc
	acquireCount            *prometheus.Desc
	acquireDuration         *prometheus.Desc
	emptyAcquireCount       *prometheus.Desc
	canceledAcquireCount    *prometheus.Desc
	newConnsCount           *prometheus.Desc
	maxLifetimeDestroyCount *prometheus.Desc
	maxIdleDestroyCount     *prometheus.Desc
}

// NewPoolCollector builds a collector reading from the supplied pool. Register
// it on a Prometheus registry (typically prometheus.MustRegister so it is
// exposed by the same /metrics handler as the rest of the server's metrics).
func NewPoolCollector(pool *pgxpool.Pool) *PoolCollector {
	return newPoolCollector(pool)
}

// newPoolCollector is the interface-typed constructor the exported wrapper and
// the tests share.
func newPoolCollector(pool poolStatter) *PoolCollector {
	const namespace = "db_pool"

	return &PoolCollector{
		pool: pool,
		acquiredConns: prometheus.NewDesc(
			namespace+"_acquired_connections",
			"Number of currently acquired (in-use) connections in the pool.",
			nil, nil,
		),
		idleConns: prometheus.NewDesc(
			namespace+"_idle_connections",
			"Number of currently idle connections in the pool.",
			nil, nil,
		),
		totalConns: prometheus.NewDesc(
			namespace+"_total_connections",
			"Total number of connections currently in the pool (acquired plus idle plus constructing).",
			nil, nil,
		),
		maxConns: prometheus.NewDesc(
			namespace+"_max_connections",
			"Maximum number of connections the pool is configured to allow.",
			nil, nil,
		),
		constructingConns: prometheus.NewDesc(
			namespace+"_constructing_connections",
			"Number of connections currently being established.",
			nil, nil,
		),
		acquireCount: prometheus.NewDesc(
			namespace+"_acquire_total",
			"Cumulative count of successful connection acquisitions from the pool.",
			nil, nil,
		),
		acquireDuration: prometheus.NewDesc(
			namespace+"_acquire_duration_seconds",
			"Cumulative time spent (in seconds) acquiring connections, including acquisitions satisfied instantly from an idle connection.",
			nil, nil,
		),
		emptyAcquireCount: prometheus.NewDesc(
			namespace+"_empty_acquire_total",
			"Cumulative count of acquisitions that had to wait because the pool was empty; a rising value signals pool saturation.",
			nil, nil,
		),
		canceledAcquireCount: prometheus.NewDesc(
			namespace+"_canceled_acquire_total",
			"Cumulative count of acquisitions canceled by a context before a connection was obtained.",
			nil, nil,
		),
		newConnsCount: prometheus.NewDesc(
			namespace+"_new_connections_total",
			"Cumulative count of new connections opened by the pool.",
			nil, nil,
		),
		maxLifetimeDestroyCount: prometheus.NewDesc(
			namespace+"_max_lifetime_destroy_total",
			"Cumulative count of connections destroyed because they exceeded MaxConnLifetime.",
			nil, nil,
		),
		maxIdleDestroyCount: prometheus.NewDesc(
			namespace+"_max_idle_destroy_total",
			"Cumulative count of connections destroyed because they exceeded MaxConnIdleTime.",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *PoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquiredConns
	ch <- c.idleConns
	ch <- c.totalConns
	ch <- c.maxConns
	ch <- c.constructingConns
	ch <- c.acquireCount
	ch <- c.acquireDuration
	ch <- c.emptyAcquireCount
	ch <- c.canceledAcquireCount
	ch <- c.newConnsCount
	ch <- c.maxLifetimeDestroyCount
	ch <- c.maxIdleDestroyCount
}

// Collect implements prometheus.Collector. It takes a single stats snapshot per
// scrape so every metric in the scrape is internally consistent.
func (c *PoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()

	gauge := func(desc *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v)
	}
	counter := func(desc *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v)
	}

	gauge(c.acquiredConns, float64(stat.AcquiredConns()))
	gauge(c.idleConns, float64(stat.IdleConns()))
	gauge(c.totalConns, float64(stat.TotalConns()))
	gauge(c.maxConns, float64(stat.MaxConns()))
	gauge(c.constructingConns, float64(stat.ConstructingConns()))

	counter(c.acquireCount, float64(stat.AcquireCount()))
	counter(c.acquireDuration, stat.AcquireDuration().Seconds())
	counter(c.emptyAcquireCount, float64(stat.EmptyAcquireCount()))
	counter(c.canceledAcquireCount, float64(stat.CanceledAcquireCount()))
	counter(c.newConnsCount, float64(stat.NewConnsCount()))
	counter(c.maxLifetimeDestroyCount, float64(stat.MaxLifetimeDestroyCount()))
	counter(c.maxIdleDestroyCount, float64(stat.MaxIdleDestroyCount()))
}
