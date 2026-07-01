package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ReportsSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reports_builder_reports_sent_total",
		Help: "Total number of successfully sent reports.",
	})

	ReportsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reports_builder_reports_failed_total",
		Help: "Total number of report build/send failures.",
	})

	FlatsInReport = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "reports_builder_flats_in_report",
		Help:    "Number of flats included in each sent report.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12),
	})
)

func init() {
	prometheus.MustRegister(ReportsSent, ReportsFailed, FlatsInReport)
}
