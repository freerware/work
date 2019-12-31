all: bins

clean:
	@GO111MODULE=on go clean -x

bins:
	@GO111MODULE=on go build

test: bins
	@#v1 + v2
	@GO111MODULE=on go test -v -covermode=count -coverprofile=coverage.out ./ ./v2

mocks:
	#v1
	@mockgen -source=data_mapper.go -destination=internal/mock/data_mapper.go -package=mock -mock_names=DataMapper=DataMapper
	@mockgen -source=sql_data_mapper.go -destination=internal/mock/sql_data_mapper.go -package=mock -mock_names=SQLDataMapper=SQLDataMapper
	#v2
	@mockgen -source=v2/data_mapper.go -destination=v2/internal/mock/data_mapper.go -package=mock -mock_names=DataMapper=DataMapper
	@mockgen -source=v2/sql_data_mapper.go -destination=v2/internal/mock/sql_data_mapper.go -package=mock -mock_names=SQLDataMapper=SQLDataMapper
