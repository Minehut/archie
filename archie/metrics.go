package archie

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"regexp"
	"strconv"
	"strings"
)

var (
	subSystem = "archie"

	messagesProcessedCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subSystem,
			Name:      "messages_processed_count",
			Help:      "count of messages processed by state",
		},
		[]string{"state", "error", "code", "event", "eventType"},
	)

	// transfer
	messagesTransferDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_transfer_duration",
		Help:      "a histogram of file transfer duration in seconds",
		Buckets:   []float64{3, 5, 10, 30, 60, 120, 240, 300, 600, 900, 1800, 3600},
	})
	messagesTransferRateMetric = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_transfer_rate",
		Help:      "a histogram of file transfer speed in kbytes/second",
		Buckets:   []float64{500, 1_000, 5_000, 10_000, 12_000, 15_000, 20_000, 25_000, 30_000, 50_000, 70_000},
	})
	messagesTransferSizeMetric = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_transfer_size",
		Help:      "a histogram of file transfer size in kbytes",
		Buckets:   []float64{1_000, 10_000, 50_000, 100_000, 500_000, 1_000_000, 5_000_000, 10_000_000, 20_000_000, 50_000_000},
	})
	messagesTransferNumDeliveredMetric = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_transfer_delivered_count",
		Help:      "a histogram of the number of times a jetstream message was delivered before it was successful",
		Buckets:   []float64{1, 2, 3, 5, 10, 20, 30, 40},
	})
	messagesTransferQueueDurationMetric = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_transfer_queue_duration",
		Help:      "a histogram of the duration of time a message spent waiting and retrying in the queue in seconds",
		Buckets:   []float64{10, 30, 60, 120, 240, 300, 600, 900, 1800, 3600, 7200, 21_600, 43_200},
	})

	// delete
	messagesDeleteDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_delete_duration",
		Help:      "a histogram of file delete duration in seconds",
		Buckets:   []float64{1, 2, 3, 5, 10, 30, 60},
	})
	messagesDeleteNumDeliveredMetric = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_delete_delivered_count",
		Help:      "a histogram of the number of times a delete jetstream message was delivered before it was successful",
		Buckets:   []float64{1, 2, 3, 5, 10, 20, 30, 40},
	})
	messagesDeleteQueueDurationMetric = promauto.NewHistogram(prometheus.HistogramOpts{
		Subsystem: subSystem,
		Name:      "messages_delete_queue_duration",
		Help:      "a histogram of the duration of time a delete jetstream message spent waiting and retrying in the queue in seconds",
		Buckets:   []float64{10, 30, 60, 120, 240, 300, 600, 900, 1800, 3600, 7200, 21_600, 43_200},
	})
)

func (a *Archiver) countMessagesProcessedMetric(state string, error string, code string, event string, eventType string) {
	messagesProcessedCount.WithLabelValues(state, error, code, event, eventType).Inc()
}
func (a *Archiver) observeMessagesTransferDurationMetric(seconds float64) {
	messagesTransferDuration.Observe(seconds)
}
func (a *Archiver) observeMessagesTransferRateMetric(bytes float64) {
	kBytes := bytes / 1000
	messagesTransferRateMetric.Observe(kBytes)
}
func (a *Archiver) observeMessagesTransferSizeMetric(bytes float64) {
	kBytes := bytes / 1000
	messagesTransferSizeMetric.Observe(kBytes)
}
func (a *Archiver) observeMessagesTransferNumDeliveredMetric(count float64) {
	messagesTransferNumDeliveredMetric.Observe(count)
}
func (a *Archiver) observeMessagesTransferQueueDurationMetric(seconds float64) {
	messagesTransferQueueDurationMetric.Observe(seconds)
}
func (a *Archiver) observeMessagesDeleteDurationMetric(seconds float64) {
	messagesDeleteDuration.Observe(seconds)
}
func (a *Archiver) observeMessagesDeleteNumDeliveredMetric(count float64) {
	messagesDeleteNumDeliveredMetric.Observe(count)
}
func (a *Archiver) observeMessagesDeleteQueueDurationMetric(seconds float64) {
	messagesDeleteQueueDurationMetric.Observe(seconds)
}

func (a *Archiver) cleanupAndCountMessagesProcessedMetric(state string, error string, code string, event string, eventType string) {
	// remove any URLs from the error output
	urlRegex := regexp.MustCompile(`((")?http(s)?://[\w.\-/?=&:"_]+)`)
	// remove any IPs from the error output
	ipRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(:\d{1,5})?)`)

	stripPrefixes := []string{"Post", "Put", "Head", "Get"}
	for _, prefix := range stripPrefixes {
		if strings.HasPrefix(error, prefix) {
			error = urlRegex.ReplaceAllString(error, "-")
			error = ipRegex.ReplaceAllString(error, "-")
			break
		}
	}

	// strip incident id and carriage returns from embedded error msg
	if strings.Contains(error, "incident id") {
		incidentRegex := regexp.MustCompile(`(incident id )(\w+-\w+)`)
		returnRegex := regexp.MustCompile(`\n\s?`)

		error = incidentRegex.ReplaceAllString(error, "$1 -")
		error = returnRegex.ReplaceAllString(error, "")

		unquotedError, err := strconv.Unquote(fmt.Sprintf(`"%s"`, error))
		if err == nil {
			error = unquotedError
		}
	}

	a.countMessagesProcessedMetric(state, error, code, event, eventType)
}
