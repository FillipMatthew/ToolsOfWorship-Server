package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tow_http_requests_total",
		Help: "Total HTTP requests by method, pattern, and status code.",
	}, []string{"method", "pattern", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tow_http_request_duration_seconds",
		Help:    "HTTP request latency by method and pattern.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "pattern"})

	httpErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tow_http_errors_total",
		Help: "Total HTTP errors by method, pattern, and error code.",
	}, []string{"method", "pattern", "error_code"})
)

// statusResponseWriter wraps http.ResponseWriter to capture the written status code.
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (sw *statusResponseWriter) WriteHeader(code int) {
	if !sw.written {
		sw.statusCode = code
		sw.written = true
		sw.ResponseWriter.WriteHeader(code)
	}
}

func (sw *statusResponseWriter) Write(b []byte) (int, error) {
	if !sw.written {
		sw.WriteHeader(http.StatusOK)
	}
	return sw.ResponseWriter.Write(b)
}

// MetricsMiddleware records per-request Prometheus metrics.
func MetricsMiddleware(method, pattern string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()

		err := h.ServeHTTP(sw, r)

		httpRequestsTotal.WithLabelValues(method, pattern, strconv.Itoa(sw.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(method, pattern).Observe(time.Since(start).Seconds())

		if err != nil {
			errCode := ""
			var apiErr *Error
			if errors.As(err, &apiErr) {
				errCode = apiErr.ErrorCode
			}
			httpErrorsTotal.WithLabelValues(method, pattern, errCode).Inc()
		}

		return err
	})
}

// dbStatsCollector exposes sql.DBStats fields as Prometheus metrics.
type dbStatsCollector struct {
	db             *sql.DB
	openConns      *prometheus.Desc
	inUseConns     *prometheus.Desc
	idleConns      *prometheus.Desc
	waitCount      *prometheus.Desc
	maxOpenConns   *prometheus.Desc
}

// NewDBStatsCollector returns a prometheus.Collector for sql.DB pool statistics.
func NewDBStatsCollector(db *sql.DB) prometheus.Collector {
	return &dbStatsCollector{
		db: db,
		openConns: prometheus.NewDesc(
			"tow_db_open_connections",
			"Current number of open DB connections.",
			nil, nil,
		),
		inUseConns: prometheus.NewDesc(
			"tow_db_in_use_connections",
			"Current number of DB connections in use.",
			nil, nil,
		),
		idleConns: prometheus.NewDesc(
			"tow_db_idle_connections",
			"Current number of idle DB connections.",
			nil, nil,
		),
		waitCount: prometheus.NewDesc(
			"tow_db_wait_count_total",
			"Total number of times a goroutine waited for a DB connection.",
			nil, nil,
		),
		maxOpenConns: prometheus.NewDesc(
			"tow_db_max_open_connections",
			"Maximum number of open DB connections allowed.",
			nil, nil,
		),
	}
}

func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.openConns
	ch <- c.inUseConns
	ch <- c.idleConns
	ch <- c.waitCount
	ch <- c.maxOpenConns
}

func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.db.Stats()
	ch <- prometheus.MustNewConstMetric(c.openConns, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(c.inUseConns, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.GaugeValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.maxOpenConns, prometheus.GaugeValue, float64(stats.MaxOpenConnections))
}
