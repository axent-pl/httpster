package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type RequestDefinition struct {
	ID      string            `json:"id"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type RequestMetrics struct {
	ID                   string        `json:"id"`
	StartTime            time.Time     `json:"start"`
	Duration             time.Duration `json:"duration"`
	DurationMilliseconds int64         `json:"durationMilliseconds"`
	Error                string        `json:"error"`
}

type TestMetrics struct {
	requestMetricsMux sync.Mutex
	requestMetrics    []RequestMetrics
}

func (rd *RequestDefinition) PrepareRequest() (*http.Request, error) {
	req, err := http.NewRequest(rd.Method, rd.URL, strings.NewReader(rd.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range rd.Headers {
		req.Header.Add(k, v)
	}
	return req, nil
}

func (rm *RequestMetrics) ToCSVRow() []string {
	return []string{
		rm.ID,
		rm.StartTime.Format(time.RFC3339),
		rm.Duration.String(),
		fmt.Sprintf("%d", rm.DurationMilliseconds),
		rm.Error,
	}
}

func GetRequestMetricsCSVHeader() []string {
	return []string{"ID", "Start Time", "Duration", "Duration Milliseconds", "Error"}
}

func (tm *TestMetrics) ToCSV() (string, error) {
	tm.requestMetricsMux.Lock()
	defer tm.requestMetricsMux.Unlock()

	var b bytes.Buffer
	writer := csv.NewWriter(&b)

	if err := writer.Write(GetRequestMetricsCSVHeader()); err != nil {
		return "", fmt.Errorf("writing header to CSV failed: %v", err)
	}

	for _, metric := range tm.requestMetrics {
		record := metric.ToCSVRow()
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("writing record to CSV failed: %v", err)
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("finalizing CSV failed: %v", err)
	}

	return b.String(), nil
}

func MakeRequest(client *http.Client, requestDefinition *RequestDefinition) (RequestMetrics, error) {
	var requestMetrics RequestMetrics = RequestMetrics{ID: requestDefinition.ID}
	req, err := requestDefinition.PrepareRequest()
	if err != nil {
		requestMetrics.Error = fmt.Sprint(err)
		return requestMetrics, err
	}
	requestMetrics.StartTime = time.Now()
	resp, err := client.Do(req)
	if err != nil {
		requestMetrics.Error = fmt.Sprint(err)
		return requestMetrics, err
	}
	resp.Body.Close()
	requestMetrics.Duration = time.Since(requestMetrics.StartTime)
	requestMetrics.DurationMilliseconds = requestMetrics.Duration.Milliseconds()
	return requestMetrics, nil
}

func RunTests(requestDefinitions []*RequestDefinition, threads int, duration time.Duration) (*TestMetrics, error) {
	var wg sync.WaitGroup
	var metrics TestMetrics = TestMetrics{}
	var stopTime time.Time = time.Now().Add(duration)

	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go func(i int) {
			defer wg.Done()
			var threadMetrics []RequestMetrics = make([]RequestMetrics, 0)
			client := &http.Client{}
			for time.Now().Before(stopTime) {
				for _, reqDefinition := range requestDefinitions {
					requestMetrics, _ := MakeRequest(client, reqDefinition)
					threadMetrics = append(threadMetrics, requestMetrics)
				}
			}
			metrics.requestMetricsMux.Lock()
			metrics.requestMetrics = append(metrics.requestMetrics, threadMetrics...)
			metrics.requestMetricsMux.Unlock()
		}(i)
	}
	wg.Wait()

	return &metrics, nil
}

func main() {
	var threads int
	var durationString string
	var requestDefinitions []*RequestDefinition

	flag.IntVar(&threads, "threads", 1, "Number of threads")
	flag.StringVar(&durationString, "duration", "10s", "Duration to run the test (e.g., 10s, 1m, 2h)")
	flag.Parse()

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(stdin, &requestDefinitions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling from stdin: %v\n", err)
		os.Exit(1)
	}

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing duration: %v", err)
		os.Exit(1)
	}

	metrics, _ := RunTests(requestDefinitions, threads, duration)
	metricsCSV, err := metrics.ToCSV()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting test metrics to CSV: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(metricsCSV)
}
