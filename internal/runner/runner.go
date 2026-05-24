package runner

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	URL         string
	Requests    int
	Concurrency int
	Method      string
	TimeoutSec  int
}

type Stats struct {
	Total    int     `json:"total"`
	Success  int     `json:"success"`
	Failed   int     `json:"failed"`
	TotalSec float64 `json:"total_time_s"`
	RPS      float64 `json:"rps"`
	MinMS    float64 `json:"min_ms"`
	MeanMS   float64 `json:"mean_ms"`
	P50MS    float64 `json:"p50_ms"`
	P90MS    float64 `json:"p90_ms"`
	P99MS    float64 `json:"p99_ms"`
	MaxMS    float64 `json:"max_ms"`
}

type result struct {
	latency time.Duration
	status  int
	err     error
}

func Run(cfg Config) (Stats, error) {
	jobs := make(chan struct{}, cfg.Requests)
	results := make(chan result, cfg.Requests)
	done := make(chan struct{})
	var completed atomic.Int64

	for i := 0; i < cfg.Requests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	client := &http.Client{
		Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
	}

	// progress bar goroutine — updates every 100ms
	var progressWg sync.WaitGroup
	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				printProgress(int(completed.Load()), cfg.Requests)
				fmt.Println()
				return
			case <-ticker.C:
				printProgress(int(completed.Load()), cfg.Requests)
			}
		}
	}()

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				results <- doRequest(client, cfg.Method, cfg.URL)
				completed.Add(1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	close(done)
	progressWg.Wait()
	close(results)

	return buildStats(cfg.Requests, elapsed, results), nil
}

func doRequest(client *http.Client, method, url string) result {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return result{err: err}
	}

	t := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(t)

	if err != nil {
		return result{latency: latency, err: err}
	}
	resp.Body.Close()
	return result{latency: latency, status: resp.StatusCode}
}

func buildStats(total int, elapsed time.Duration, results <-chan result) Stats {
	var latencies []float64
	success, failed := 0, 0

	for r := range results {
		ms := float64(r.latency.Microseconds()) / 1000.0
		latencies = append(latencies, ms)
		if r.err != nil || r.status >= 400 {
			failed++
		} else {
			success++
		}
	}

	sort.Float64s(latencies)

	s := Stats{
		Total:    total,
		Success:  success,
		Failed:   failed,
		TotalSec: elapsed.Seconds(),
		RPS:      float64(total) / elapsed.Seconds(),
	}

	if len(latencies) > 0 {
		s.MinMS = latencies[0]
		s.MaxMS = latencies[len(latencies)-1]
		s.MeanMS = mean(latencies)
		s.P50MS = percentile(latencies, 50)
		s.P90MS = percentile(latencies, 90)
		s.P99MS = percentile(latencies, 99)
	}

	return s
}

func printProgress(done, total int) {
	const width = 30
	filled := 0
	if total > 0 {
		filled = done * width / total
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}
	fmt.Printf("\r  [%s] %3d%% (%d/%d)", bar, pct, done, total)
}

func mean(vals []float64) float64 {
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100)
	return sorted[idx]
}
