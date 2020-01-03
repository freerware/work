all: bins

clean:
	go clean -x
	rm -r --force vendor/

bins:
	dep ensure -v
	go build

test: bins
	go test -v -covermode=count -coverprofile=coverage.out

mocks:
	mockgen -source=data_mapper.go -destination=internal/mock/data_mapper.go -package=mock -mock_names=DataMapper=DataMapper
	mockgen -source=sql_data_mapper.go -destination=internal/mock/sql_data_mapper.go -package=mock -mock_names=SQLDataMapper=SQLDataMapper
