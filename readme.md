# HTTPSTER

A simple load test tool that executes multiple requests with any method (GET, POST, PUT, ...) and returns raw data to perform detailed analysis.

## Usage

```shell
# build
make build

# check for options
bin/httpster --help

# run
echo '[{"id":"GET/","url":"http://httpbin.org", "method":"GET"},{"id":"GET/anything","url":"http://httpbin.org/anything", "method":"GET"}]' | bin/httpster -duration=10s -threads=2 1> out/data.csv
```