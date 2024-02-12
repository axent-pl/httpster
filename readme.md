# HTTPSTER

A simple load test tool that executes multiple requests with any method (GET, POST, PUT, ...) and returns raw data to perform detailed analysis.

## Usage

```shell
# build
make build

# check for options
bin/httpster --help

# run
cat example.json | bin/httpster -duration=30s -threads=4 1> out/data.csv
```

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