// This is a simple Go program that executes an AppleScript to get the number
// of emails in the inbox of each account in Mail.app and pushes the metrics
// to a Prometheus Push Gateway.
// Usage: go run main.go -script counter_by_inbox.scpt -push-gateway localhost:9091 -metric-name email_count -job-name email_count
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func run() error {
	scriptPath := flag.String("script", "counter_by_inbox.scpt", "Path to AppleScript script")
	pushGateway := flag.String("push-gateway", "localhost:9091", "Address of Prometheus Push Gateway")
	metricName := flag.String("metric-name", "mailbox_inbox_count", "Name of the metric to push")
	jobName := flag.String("job-name", "mailbox", "Name of the job to push")
	flag.Parse()

	out, err := runScript("/usr/bin/osascript", *scriptPath)
	if err != nil {
		log.Printf("Output: %s", out)
		return fmt.Errorf("failed to execute AppleScript: %w", err)
	}

	metrics, err := parseMetris(out)
	if err != nil {
		return fmt.Errorf("failed to parse metrics: %w", err)
	}

	if err := pushMetrics(
		metrics,
		*pushGateway,
		*metricName,
		*jobName,
	); err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}

	log.Printf("Pushed %d metrics", len(metrics))

	return nil
}

func runScript(name string, arg ...string) ([]byte, error) {
	// execute AppleScript to get Mailbox inbox counts
	cmd := exec.Command(name, arg...)
	return cmd.CombinedOutput()
}

func parseMetris(b []byte) (map[string]int, error) {
	metrics := make(map[string]int)
	lines := splitLines(b)
	for _, line := range lines {
		// parse line
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		name := fields[0]
		count, err := strconv.Atoi(strings.TrimSpace(fields[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid count for %s: %s", name, fields[1])
		}
		metrics[name] = count
	}
	return metrics, nil
}

func splitLines(b []byte) []string {
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := sc.Text()
		if len(line) > 0 {
			lines = append(lines, sc.Text())
		}
	}
	return lines
}

func pushMetrics(
	metrics map[string]int,
	pushGateway string,
	metricName string,
	jobName string,
) error {
	counter := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricName,
		Help: "Number of emails in the inbox",
	})

	for name, count := range metrics {
		counter.Set(float64(count))
		if err := push.New(pushGateway, jobName).
			Collector(counter).
			Grouping("account", name).
			Push(); err != nil {
			return fmt.Errorf("failed to push metrics: %w", err)
		}
	}
	return nil
}
