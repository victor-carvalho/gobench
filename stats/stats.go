package stats

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"time"

	"github.com/victor-carvalho/gobench/config"
	"github.com/victor-carvalho/gobench/requester"
)

type Percentiles struct {
	p50 time.Duration
	p75 time.Duration
	p90 time.Duration
	p95 time.Duration
	p99 time.Duration
}

type LatencyStats struct {
	Average     time.Duration
	Minimum     time.Duration
	Maximum     time.Duration
	Percentiles Percentiles
}

type Stats struct {
	RequestStats    LatencyStats
	ServerStats     LatencyStats
	TotalStats      LatencyStats
	TotalTime       time.Duration
	TotalErrors     int
	TotalTimeouts   int
	TotalMatches    int
	TotalNonMatches int
	TotalRequests   int
}

func showTime(millis float64) string {
	if millis < 1_000 {
		return fmt.Sprintf("%fms", millis)
	}
	return fmt.Sprintf("%fs", float64(millis)/1000)
}

func (s Stats) Show() {
	fmt.Print("Summary:\n")
	fmt.Printf("- Total requests: %d\n", s.TotalRequests)
	fmt.Printf("- Total errors: %d (%.2f%%)\n", s.TotalErrors, 100*float64(s.TotalErrors)/float64(s.TotalRequests))
	fmt.Printf("- Total timeouts: %d (%.2f%%)\n", s.TotalTimeouts, 100*float64(s.TotalTimeouts)/float64(s.TotalRequests))
	fmt.Printf("- Total requests match status: %d (%.2f%%)\n", s.TotalMatches, 100*float64(s.TotalMatches)/float64(s.TotalRequests))
	fmt.Printf("- Total requests didn't match status: %d (%.2f%%)\n", s.TotalNonMatches, 100*float64(s.TotalNonMatches)/float64(s.TotalRequests))
	fmt.Printf("- Total time spent during bench: %s\n", s.TotalTime)
	fmt.Printf("\n")
	fmt.Printf("Request latency stats (time to connect to the server):")
	fmt.Printf("- Minimum latency: %s\n", s.RequestStats.Minimum)
	fmt.Printf("- Maximum latency: %s\n", s.RequestStats.Maximum)
	fmt.Printf("- Average latency: %s\n", s.RequestStats.Average)
	fmt.Printf("- Median: %s\n", s.RequestStats.Percentiles.p50)
	fmt.Printf("- Percentil 75: %s\n", s.RequestStats.Percentiles.p75)
	fmt.Printf("- Percentil 90: %s\n", s.RequestStats.Percentiles.p90)
	fmt.Printf("- Percentil 95: %s\n", s.RequestStats.Percentiles.p95)
	fmt.Printf("- Percentil 99: %s\n", s.RequestStats.Percentiles.p99)
	fmt.Printf("\n")
	fmt.Printf("Server latency stats (time to receive response from server after connected):")
	fmt.Printf("- Minimum latency: %s\n", s.ServerStats.Minimum)
	fmt.Printf("- Maximum latency: %s\n", s.ServerStats.Maximum)
	fmt.Printf("- Average latency: %s\n", s.ServerStats.Average)
	fmt.Printf("- Median: %s\n", s.ServerStats.Percentiles.p50)
	fmt.Printf("- Percentil 75: %s\n", s.ServerStats.Percentiles.p75)
	fmt.Printf("- Percentil 90: %s\n", s.ServerStats.Percentiles.p90)
	fmt.Printf("- Percentil 95: %s\n", s.ServerStats.Percentiles.p95)
	fmt.Printf("- Percentil 99: %s\n", s.ServerStats.Percentiles.p99)
	fmt.Printf("\n")
	fmt.Printf("Total latency stats (time to connect to the server and receive response):")
	fmt.Printf("- Minimum latency: %s\n", s.TotalStats.Minimum)
	fmt.Printf("- Maximum latency: %s\n", s.TotalStats.Maximum)
	fmt.Printf("- Average latency: %s\n", s.TotalStats.Average)
	fmt.Printf("- Median: %s\n", s.TotalStats.Percentiles.p50)
	fmt.Printf("- Percentil 75: %s\n", s.TotalStats.Percentiles.p75)
	fmt.Printf("- Percentil 90: %s\n", s.TotalStats.Percentiles.p90)
	fmt.Printf("- Percentil 95: %s\n", s.TotalStats.Percentiles.p95)
	fmt.Printf("- Percentil 99: %s\n", s.TotalStats.Percentiles.p99)
	fmt.Printf("\n")
}

func countMatches(pattern *regexp.Regexp, completed []requester.Response) (matches int, nonMatches int) {
	matches = 0
	nonMatches = 0
	for _, response := range completed {
		if pattern.MatchString(response.Status) {
			matches++
		} else {
			nonMatches++
		}
	}
	return matches, nonMatches
}

func splitResponses(allResponses []requester.Response) (completed []requester.Response, timeouts []requester.Response, errors []requester.Response) {
	completed = make([]requester.Response, 0)
	timeouts = make([]requester.Response, 0)
	errors = make([]requester.Response, 0)

	for _, response := range allResponses {
		if response.Timeout {
			timeouts = append(timeouts, response)
		} else if response.Error {
			errors = append(timeouts, response)
		} else {
			completed = append(completed, response)
		}
	}

	return completed, timeouts, errors
}

func computePercentile(p int, latencies []time.Duration) time.Duration {
	n := len(latencies)
	if n == 0 {
		return time.Duration(0)
	}
	index := (float64(p) * float64(n-1) / 100) + 1
	intIndex := int64(index)
	if index == float64(int64(index)) {
		return latencies[int64(index)-1]
	}
	delta := index - float64(intIndex)
	base := latencies[intIndex-1]
	next := latencies[intIndex]
	return base + time.Duration(float64(next-base)*delta)
}

func computePercentiles(latencies []time.Duration) Percentiles {
	return Percentiles{
		p50: computePercentile(50, latencies),
		p75: computePercentile(75, latencies),
		p90: computePercentile(90, latencies),
		p95: computePercentile(95, latencies),
		p99: computePercentile(99, latencies),
	}
}

func computeLatencyStats(completed []requester.Response, latencyKey func(r *requester.Response) time.Duration) LatencyStats {
	latencies := make([]time.Duration, 0, len(completed))
	for _, response := range completed {
		latencies = append(latencies, latencyKey(&response))
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var stats LatencyStats
	minimum := time.Duration(math.MaxInt64)
	maximum := time.Duration(math.MinInt64)
	total := time.Duration(0)
	for _, latency := range latencies {
		if latency < minimum {
			minimum = latency
		}
		if latency > maximum {
			maximum = latency
		}
		total += latency
	}

	stats.Minimum = minimum
	stats.Maximum = maximum
	stats.Average = total / time.Duration(len(latencies))
	stats.Percentiles = computePercentiles(latencies)

	return stats
}

func computeStats(pattern *regexp.Regexp, totalDuration time.Duration, allResponses []requester.Response) Stats {
	var stats Stats
	stats.TotalTime = totalDuration
	stats.TotalRequests = len(allResponses)

	if stats.TotalRequests == 0 {
		return stats
	}

	completed, timeouts, errors := splitResponses(allResponses)
	stats.TotalMatches, stats.TotalNonMatches = countMatches(pattern, completed)
	stats.TotalTimeouts = len(timeouts)
	stats.TotalErrors = len(errors)
	stats.RequestStats = computeLatencyStats(completed, func(r *requester.Response) time.Duration { return r.RequestLatency() })
	stats.ServerStats = computeLatencyStats(completed, func(r *requester.Response) time.Duration { return r.ServerLatency() })
	stats.TotalStats = computeLatencyStats(completed, func(r *requester.Response) time.Duration { return r.TotalLatency() })
	return stats
}

func CollectStats(ctx context.Context, cfg config.Config, inputCh chan requester.Response) Stats {
	start := time.Now()
	allResponses := make([]requester.Response, 0, cfg.MaxRequests)
	for i := 0; i < cfg.MaxRequests; i++ {
		select {
		case response, ok := <-inputCh:
			if !ok {
				break
			}
			allResponses = append(allResponses, response)
		case <-ctx.Done():
			break
		}
	}

	return computeStats(cfg.StatusPattern, time.Since(start), allResponses)
}
