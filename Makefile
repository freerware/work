all: bins

clean:
	go clean -x
	rm -r --force vendor/

bins:
	dep ensure -v
	go build

test: bins
	go test -v -covermode=count -coverprofile=coverage.out
