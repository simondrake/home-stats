DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: build-darwin
build-darwin:
	CGO_ENABLED=0 go build \
		-o ./builds/home-stats-darwin-$(DATE) \
		-ldflags "\
			-X main.date=$(DATE) \
		" ./cmd/home-stats/main.go

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-o ./builds/home-stats-$(DATE) \
		-ldflags "\
			-X main.date=$(DATE) \
		" ./cmd/home-stats/main.go

.PHONY: test
test:
	go test ./...

.PHONY: run
run:
	go run cmd/home-stats/main.go
