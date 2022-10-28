BIN = ./bin/opscli

format:
	go mod tidy
	go mod vendor
	go fmt $(go list ./... | grep -v /vendor/)
	go test $(go list ./... | grep -v /vendor/)

run:
	go run main.go

binary:
	go build -ldflags "-w -s" -o $(BIN) ./main.go
