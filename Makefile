BIN             = ./bin/opscli

format:
	go mod tidy
	go mod vendor
	go fmt $(go list ./... | grep -v /vendor/)
	go test $(go list ./... | grep -v /vendor/)

run:
	go run main.go

binary:
	go build -ldflags "-w -s" -o $(BIN) ./main.go

tag:
	git tag -d latest || true
	git push -d origin latest || true
	git tag latest
	git push origin latest
