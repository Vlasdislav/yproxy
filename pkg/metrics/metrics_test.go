package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func histogramState(t *testing.T, observer prometheus.Observer) (uint64, float64) {
	t.Helper()

	histogram, ok := observer.(prometheus.Histogram)
	if !ok {
		t.Fatal("metric is not a histogram")
	}

	metric := &dto.Metric{}
	if err := histogram.Write(metric); err != nil {
		t.Fatalf("write histogram metric: %v", err)
	}

	return metric.GetHistogram().GetSampleCount(), metric.GetHistogram().GetSampleSum()
}

func TestStoreLatencyAndSizeInfoRecordsSeconds(t *testing.T) {
	const source = "LIMIT_READ"

	latencyBeforeCount, latencyBeforeSum := histogramState(
		t, HistogramLatencyVec.WithLabelValues(source))
	sizeBeforeCount, sizeBeforeSum := histogramState(
		t, HistogramSizeVec.WithLabelValues(source))

	StoreLatencyAndSizeInfo(source, 1024, 250*time.Millisecond)

	latencyAfterCount, latencyAfterSum := histogramState(
		t, HistogramLatencyVec.WithLabelValues(source))
	sizeAfterCount, sizeAfterSum := histogramState(
		t, HistogramSizeVec.WithLabelValues(source))

	if latencyAfterCount != latencyBeforeCount+1 {
		t.Fatalf("latency observations: got %d, want %d", latencyAfterCount, latencyBeforeCount+1)
	}
	if math.Abs(latencyAfterSum-latencyBeforeSum-0.25) > 1e-9 {
		t.Fatalf("latency sum: got %f, want increment 0.25", latencyAfterSum-latencyBeforeSum)
	}
	if sizeAfterCount != sizeBeforeCount+1 {
		t.Fatalf("size observations: got %d, want %d", sizeAfterCount, sizeBeforeCount+1)
	}
	if sizeAfterSum-sizeBeforeSum != 1024 {
		t.Fatalf("size sum: got %f, want increment 1024", sizeAfterSum-sizeBeforeSum)
	}
}
