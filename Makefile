.PHONY: test build build-erx clean

test:
	go test ./...

build:
	go build -trimpath -ldflags="-s -w" -o bin/edge-sensor ./cmd/edge-sensor

build-erx:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -trimpath -ldflags="-s -w" -o dist/edge-sensor-linux-mipsle ./cmd/edge-sensor

clean:
	rm -rf bin dist