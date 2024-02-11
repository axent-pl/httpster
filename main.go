package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type RequestDefinition struct {
	ID      string            `json:"id"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
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
	req, err := http.NewRequest(rd.Method, rd.URL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range rd.Headers {
		req.Header.Add(k, v)
	}
	return req, nil
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

func RunTests(requestDefinitions []*RequestDefinition, threads int, duration time.Duration) ([]RequestMetrics, error) {
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

	return metrics.requestMetrics, nil
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
		fmt.Printf("Error reading from stdin: %v\n", err)
		return
	}

	err = json.Unmarshal(stdin, &requestDefinitions)
	if err != nil {
		fmt.Printf("Error unmarshaling from stdin: %v\n", err)
		return
	}

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		fmt.Println("Error parsing duration:", err)
		return
	}

	metrics, _ := RunTests(requestDefinitions, threads, duration)
	metricsJson, err := json.Marshal(metrics)
	if err != nil {
		fmt.Printf("Error marshalling metrics: %v\n", err)
		return
	}
	fmt.Print(string(metricsJson))
}
