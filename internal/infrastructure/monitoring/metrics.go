package monitoring

import "github.com/prometheus/client_golang/prometheus"

var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP-запросов",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP-запросов в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	CreatedPVZCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_pvz_total",
			Help: "Количество созданных ПВЗ",
		},
	)

	CreatedReceptionsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_receptions_total",
			Help: "Количество созданных приемок",
		},
	)

	AddedProductsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_products_total",
			Help: "Количество добавленных товаров",
		},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(
		TotalRequests,
		RequestDuration,
		CreatedPVZCounter,
		CreatedReceptionsCounter,
		AddedProductsCounter,
	)
}
