BIN             = ./bin/opscli

format:
	go fmt $(go list ./... | grep -v /vendor/)
	go test $(go list ./... | grep -v /vendor/)

vendor:
	go mod tidy
	go mod vendor

run:
	go run main.go

binary:
	go build -ldflags "-w -s" -o $(BIN) ./main.go

tag:
	git tag -d v1.0.0 || true
	git push -d origin v1.0.0 || true
	git tag v1.0.0
	git push origin v1.0.0

clearhistory:
	git checkout main
	git checkout --orphan new_main
	git add -A
	git commit -m "init"
	git branch -D main
	git branch -m main
	git push -f origin main
