package main

import (
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

type RequestMetrics struct {
	ID                   string        `json:"id"`
	StartTime            time.Time     `json:"start"`
	Duration             time.Duration `json:"duration"`
	DurationMilliseconds int64         `json:"durationMilliseconds"`
	Status               string        `json:"status"`
	StatusCode           int           `json:"statusCode"`
	Error                string        `json:"error"`
}

func (rm *RequestMetrics) ToCSVRow() []string {
	return []string{
		rm.ID,
		rm.StartTime.Format(time.RFC3339),
		rm.Duration.String(),
		fmt.Sprintf("%d", rm.DurationMilliseconds),
		rm.Status,
		fmt.Sprintf("%d", rm.StatusCode),
		rm.Error,
	}
}

func GetRequestMetricsCSVHeader() []string {
	return []string{"ID", "Start Time", "Duration", "Duration Milliseconds", "Status", "StatusCode", "Error"}
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
	requestMetrics.Status = resp.Status
	requestMetrics.StatusCode = resp.StatusCode
	requestMetrics.Duration = time.Since(requestMetrics.StartTime)
	requestMetrics.DurationMilliseconds = requestMetrics.Duration.Milliseconds()
	return requestMetrics, nil
}

func RunTests(requestDefinitions []*RequestDefinition, threads int, duration time.Duration) {
	var wg sync.WaitGroup
	var stopTime time.Time = time.Now().Add(duration)
	var writer *csv.Writer = csv.NewWriter(os.Stdout)
	var metricsChannel chan RequestMetrics = make(chan RequestMetrics, threads)

	if err := writer.Write(GetRequestMetricsCSVHeader()); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV header: %v", err)
		return
	}

	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go func(i int) {
			defer wg.Done()
			client := &http.Client{}
			for time.Now().Before(stopTime) {
				for _, reqDefinition := range requestDefinitions {
					requestMetrics, err := MakeRequest(client, reqDefinition)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error making request: %v", err)
					}
					metricsChannel <- requestMetrics
				}
			}
		}(i)
	}

	go func() {
		for requestMetrics := range metricsChannel {
			csvRow := requestMetrics.ToCSVRow()
			if err := writer.Write(csvRow); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing CSV header: %v", err)
			}
			writer.Flush()
		}
	}()

	wg.Wait()

	close(metricsChannel)
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

	RunTests(requestDefinitions, threads, duration)
}
