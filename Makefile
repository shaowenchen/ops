BIN = ./bin/opscli

format:
	go mod tidy
	go mod vendor

test:
	go test $(go list ./... | grep -v /vendor/)

cli:
	go build -ldflags "-w -s" -o $(BIN) ./cmd/cli/main.go
