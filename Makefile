APP=delayed-norifier
APP_EXECUTABLE="./out/$(APP)"

run:
	docker-compose down -v
	docker-compose build --no-cache
	docker-compose up

quality:
	make lint
	make fmt
	make vet

lint:
	golangci-lint run --enable-all

test:
	make tidy
	make vendor
	go test -v -timeout 10m ./test/ -coverprofile=coverage.out -json > report.json
	@echo "test done"

clean:
	go clean
	rm -rf out/
	rm -f coverage*.out

tools:
	go get github.com/rabbitmq/amqp091-go
	go get github.com/ilyakaznacheev/cleanenv
	go get golang.org/x/sync/errgroup
	go mod tidy
