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
df['StartTime'] = pd.to_datetime(df['StartTime'])

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
ID,StartTime,Duration_ns,ConnDuration_ns,DialDuration_ns,DNSDuration_ns,Status,StatusCode,Error
GET /,2024-02-12T12:50:58+01:00,338865000,187818875,146996875,40262209,200 OK,200,
GET /,2024-02-12T12:50:58+01:00,342826542,188442916,147601250,40228166,200 OK,200,
GET /,2024-02-12T12:50:58+01:00,343047959,187695292,146739750,40384542,200 OK,200,
GET /,2024-02-12T12:50:58+01:00,484274459,333713000,292769375,40259625,200 OK,200,
GET /anything,2024-02-12T12:50:58+01:00,296632375,147367458,145465667,1757708,200 OK,200,
GET /anything,2024-02-12T12:50:58+01:00,292864292,146642000,145832292,718333,200 OK,200,
GET /anything,2024-02-12T12:50:58+01:00,293629125,146643958,145827709,703833,200 OK,200,
GET /anything,2024-02-12T12:50:58+01:00,299385209,149145125,146590250,2280750,200 OK,200,
PUT /anything,2024-02-12T12:50:59+01:00,298021375,149354792,147107584,2104166,200 OK,200,
PUT /anything,2024-02-12T12:50:59+01:00,297629958,149104042,147283459,1743791,200 OK,200,
PUT /anything,2024-02-12T12:50:59+01:00,296908417,148088459,146979500,987750,200 OK,200,
```