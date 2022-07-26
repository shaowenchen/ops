VERSION         = latest
BIN             = ./bin/opscli

format:
	go fmt $(go list ./... | grep -v /vendor/)
	go mod tidy
	go mod vendor

run:
	go run main.go

binary:
	go build -ldflags "-w -s -X main.VERSION=$(RELEASE_TAG) -X main.BUILD_DATE=$(NOW)" -o $(BIN) ./main.go

clear:
	rm -rf ./bin/*

tag:
	git tag -d v1.0.0 || true
	git push -d origin v1.0.0 || true
	git tag v1.0.0
	git push origin v1.0.0

clearhistory:
	git checkout master
	git checkout --orphan new_master
	git add -A
	git commit -m "init"
	git branch -D master
	git branch -m master
	git push -f origin master
