APP=delayed-norifier
APP_EXECUTABLE="./out/$(APP)"

make run:
	make build
	chmod +x $(APP_EXECUTABLE)
	$(APP_EXECUTABLE)

make quality:
	make lint
	make fmt
	make vet

make lint:
	golangci-lint run --enable-all

make test:
	make tidy
	make vendor
	go test -v -timeout 10m ./test/ -coverprofile=coverage.out -json > report.json
	@echo "test done"

make clean:
	go clean
	rm -rf out/
	rm -f coverage*.out

make tools:
	go get github.com/rabbitmq/amqp091-go
	go get github.com/ilyakaznacheev/cleanenv
