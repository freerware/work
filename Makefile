all: bins

clean:
	@echo cleaning...
	@GO111MODULE=on go clean -x
	@echo done!

bins-v3:
	@echo building v3...
	@cd ./v3 && GO111MODULE=on go build github.com/freerware/work/v3 && cd ..
	@echo done!

bins-v4:
	@echo building v4...
	@cd ./v4 && GO111MODULE=on go build github.com/freerware/work/v4 && cd ..
	@echo done!

bins: bins-v3 bins-v4

tests-v3: bins
	@echo testing v3...
	@cd ./v3 && GO111MODULE=on go test -v -race -covermode=atomic -coverprofile=work.coverprofile github.com/freerware/work/v3 && cd ..
	@echo done!

tests-v4: bins
	@echo testing v3...
	@cd ./v4 && GO111MODULE=on go test -v -race -covermode=atomic -coverprofile=work.coverprofile github.com/freerware/work/v4 && cd ..
	@echo done!

tests: tests-v3 tests-v4

mocks-v3:
	@echo making v3 mocks...
	@mockgen -source=v3/data_mapper.go -destination=v3/internal/mock/data_mapper.go -package=mock -mock_names=DataMapper=DataMapper
	@mockgen -source=v3/sql_data_mapper.go -destination=v3/internal/mock/sql_data_mapper.go -package=mock -mock_names=SQLDataMapper=SQLDataMapper
	@echo done!

mocks-v4:
	@echo making v4 mocks...
	@mockgen -source=v4/data_mapper.go -destination=v4/internal/mock/data_mapper.go -package=mock -mock_names=DataMapper=DataMapper
	@echo done!

mocks: mocks-v3 mocks-v4

benchmark: bins
	@#v4
	@cd ./v4 && GO111MODULE=on go test -run XXX -bench . && cd ..

demo: bins
	@docker-compose --file ./docker/docker-compose.yaml up -d
	@open "http://localhost:3001"
	@echo demoing...
	@cd v4/internal/main && go run metrics_demo.go && cd ../../../
