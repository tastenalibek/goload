package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tastenalibek/goload/internal/report"
	"github.com/tastenalibek/goload/internal/runner"
)

var (
	requests    int
	concurrency int
	method      string
	timeout     int
	outputJSON  bool
)

var rootCmd = &cobra.Command{
	Use:   "goload <url>",
	Short: "goload — a fast HTTP load testing tool written in Go",
	Long: `goload fires concurrent HTTP requests at a target URL and reports
latency percentiles, throughput, and error rates.`,
	Args:    cobra.ExactArgs(1),
	RunE:    run,
	Example: `  goload https://example.com
  goload https://api.example.com -n 500 -c 25
  goload https://example.com -n 1000 -c 50 --json`,
}

func run(_ *cobra.Command, args []string) error {
	cfg := runner.Config{
		URL:         args[0],
		Requests:    requests,
		Concurrency: concurrency,
		Method:      method,
		TimeoutSec:  timeout,
	}

	fmt.Printf("Sending %d %s requests to %s (concurrency: %d)\n\n",
		cfg.Requests, cfg.Method, cfg.URL, cfg.Concurrency)

	stats, err := runner.Run(cfg)
	if err != nil {
		return err
	}

	if outputJSON {
		return report.JSON(stats)
	}
	report.Table(stats)
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&requests, "requests", "n", 100, "total number of requests to send")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "number of concurrent workers")
	rootCmd.Flags().StringVarP(&method, "method", "m", "GET", "HTTP method (GET, POST, PUT, ...)")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "per-request timeout in seconds")
	rootCmd.Flags().BoolVar(&outputJSON, "json", false, "output results as JSON")
}
