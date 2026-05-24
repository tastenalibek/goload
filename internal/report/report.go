package report

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/tastenalibek/goload/internal/runner"
)

func Table(s runner.Stats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	sep := "─────────────────────────────────────────"

	fmt.Fprintln(w, sep)
	fmt.Fprintln(w, " Summary")
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Total Requests\t%d\n", s.Total)
	fmt.Fprintf(w, "  Successful\t%d\n", s.Success)
	fmt.Fprintf(w, "  Failed\t%d\n", s.Failed)
	fmt.Fprintf(w, "  Total Time\t%.2fs\n", s.TotalSec)
	fmt.Fprintf(w, "  Requests/sec\t%.2f\n", s.RPS)
	fmt.Fprintln(w, sep)
	fmt.Fprintln(w, " Latency")
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Min\t%.2f ms\n", s.MinMS)
	fmt.Fprintf(w, "  Mean\t%.2f ms\n", s.MeanMS)
	fmt.Fprintf(w, "  p50\t%.2f ms\n", s.P50MS)
	fmt.Fprintf(w, "  p90\t%.2f ms\n", s.P90MS)
	fmt.Fprintf(w, "  p99\t%.2f ms\n", s.P99MS)
	fmt.Fprintf(w, "  Max\t%.2f ms\n", s.MaxMS)
	fmt.Fprintln(w, sep)

	w.Flush()
}

func JSON(s runner.Stats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}
