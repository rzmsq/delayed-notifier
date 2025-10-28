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
	go get github.com/wb-go/wbf
	go get github.com/rs/zerolog
	go get github.com/spf13/viper
	go get golang.org/x/sync/errgroup
	go get github.com/go-playground/validator/v10
	go get github.com/google/uuid
	go mod tidy
