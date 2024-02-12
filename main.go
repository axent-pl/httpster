package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
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
	ID            string        `json:"id"`
	StartTime     time.Time     `json:"start"`
	Duration      time.Duration `json:"duration"`
	Status        string        `json:"status"`
	StatusCode    int           `json:"statusCode"`
	Error         string        `json:"error"`
	ConnStartTime time.Time     `json:"connStarted"`
	ConnDuration  time.Duration `json:"connDuration"`
	DialStartTime time.Time     `json:"dialStartTime"`
	DialDuration  time.Duration `json:"dialDuration"`
	DNSStartTime  time.Time     `json:"dnsStartTime"`
	DNSDuration   time.Duration `json:"dnsDuration"`
}

func (rm *RequestMetrics) ToCSVRow() []string {
	return []string{
		rm.ID,
		rm.StartTime.Format(time.RFC3339),
		fmt.Sprintf("%d", rm.Duration.Nanoseconds()),
		fmt.Sprintf("%d", rm.ConnDuration.Nanoseconds()),
		fmt.Sprintf("%d", rm.DialDuration.Nanoseconds()),
		fmt.Sprintf("%d", rm.DNSDuration.Nanoseconds()),
		rm.Status,
		fmt.Sprintf("%d", rm.StatusCode),
		rm.Error,
	}
}

func GetRequestMetricsCSVHeader() []string {
	return []string{"ID", "StartTime", "Duration_ns", "ConnDuration_ns", "DialDuration_ns", "DNSDuration_ns", "Status", "StatusCode", "Error"}
}

func MakeRequest(requestDefinition *RequestDefinition) (RequestMetrics, error) {
	var requestMetrics RequestMetrics = RequestMetrics{ID: requestDefinition.ID}
	req, err := requestDefinition.PrepareRequest()
	if err != nil {
		requestMetrics.Error = fmt.Sprint(err)
		return requestMetrics, err
	}

	clientTrace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			requestMetrics.ConnStartTime = time.Now()
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			requestMetrics.DNSStartTime = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			requestMetrics.DNSDuration = time.Since(requestMetrics.DNSStartTime)
		},
		ConnectStart: func(network, addr string) {
			requestMetrics.DialStartTime = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			requestMetrics.DialDuration = time.Since(requestMetrics.DialStartTime)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			requestMetrics.ConnDuration = time.Since(requestMetrics.ConnStartTime)
		},
	}
	clientTraceCtx := httptrace.WithClientTrace(req.Context(), clientTrace)
	req = req.WithContext(clientTraceCtx)

	requestMetrics.StartTime = time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		requestMetrics.Error = fmt.Sprint(err)
		return requestMetrics, err
	}
	resp.Body.Close()
	requestMetrics.Status = resp.Status
	requestMetrics.StatusCode = resp.StatusCode
	requestMetrics.Duration = time.Since(requestMetrics.StartTime)
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
			for time.Now().Before(stopTime) {
				for _, reqDefinition := range requestDefinitions {
					requestMetrics, err := MakeRequest(reqDefinition)
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
