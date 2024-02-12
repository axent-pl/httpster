# HTTPster

A simple load test tool that executes multiple requests with any method (GET, POST, PUT, ...) and returns raw data to perform detailed analysis.

## Motivation

I attempted to conduct load tests to evaluate the latency introduced by some selected API Gateways.
Unfortunately, tools like `ab` and `wrk` did not provide sufficient information — offering basically mean and standard error — and their output was less than ideal to work with - formated text.
This prompted the idea to develop a simple tool that could provide raw data for more thorough analysis.

## Usage

```shell
# build
make build

# check for options
bin/httpster --help

# run
cat example.json | bin/httpster -duration=30s -threads=4 1> out/data.csv
```

Please note the `1>` as all the metrics are sent to the `stdout` whereas all the errors to `stderr`.

## Analysis

An example notebook containing analysis of the `HTTPster` output can be found [here](https://github.com/axent-pl/httpster/blob/main/analyse.ipynb).

To load the data from `HTTPster` please start with the following Python snippet:
```python
import pandas as pd

df = pd.read_csv("out/data.csv")
df['Start Time'] = pd.to_datetime(df['Start Time'])

```

From this point you can start your analysis - plot all the distributions, check for anomalies, etc.

## Input format

Request definitions are supposed to be provided via `stdin` in a `JSON` format with following schema:

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "array",
  "items": [
    {
      "type": "object",
      "properties": {
        "id": {
          "type": "string"
        },
        "url": {
          "type": "string"
        },
        "method": {
          "type": "string"
        },
        "headers": {
          "type": "object"
        },
        "body": {
          "type": "string"
        }
      },
      "required": [
        "url",
        "method"
      ]
    }
  ]
}
```

Example
```json
[
    {
        "id": "PUT /anything",
        "url": "http://httpbin/anything",
        "method": "PUT",
        "headers": {
            "Content-Type": "application/json"
        },
        "body": "{\"data\": {\"key1\":\"val1\"}}"
    }
]
```

## Output format

`HTTPster` ouutput is in `CSV` format to `stdout`. Any error encountered are outputed to `stderr`. Below is an example of the output

```csv
ID,Start Time,Duration,Duration Milliseconds,Status,StatusCode,Error
GET /,2024-02-12T11:52:37+01:00,340.005834ms,340,200 OK,200,
GET /,2024-02-12T11:52:37+01:00,342.842916ms,342,200 OK,200,
GET /,2024-02-12T11:52:37+01:00,343.309959ms,343,200 OK,200,
GET /,2024-02-12T11:52:37+01:00,343.321083ms,343,200 OK,200,
GET /anything,2024-02-12T11:52:37+01:00,298.535583ms,298,200 OK,200,
GET /anything,2024-02-12T11:52:37+01:00,297.305ms,297,200 OK,200,
GET /anything,2024-02-12T11:52:37+01:00,297.423791ms,297,200 OK,200,
GET /anything,2024-02-12T11:52:37+01:00,297.337083ms,297,200 OK,200,
PUT /anything,2024-02-12T11:52:37+01:00,300.084125ms,300,200 OK,200,
PUT /anything,2024-02-12T11:52:37+01:00,298.0525ms,298,200 OK,200,
PUT /anything,2024-02-12T11:52:37+01:00,300.5025ms,300,200 OK,200,
```