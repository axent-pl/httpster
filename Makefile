build:
	go build -o bin/httpster .

run:
	go run . -threads=2 -duration=5s