APP=delayed-norifier
APP_EXECUTABLE="./out/$(APP)"

run:
	docker-compose down -v
	docker-compose build --no-cache
	docker-compose up

quality:
	make lint ./...
	make vet ./...

lint:
	golangci-lint run ./...

tools:
	go get github.com/rabbitmq/amqp091-go
	go get github.com/wb-go/wbf
	go get github.com/rs/zerolog
	go get github.com/spf13/viper
	go get golang.org/x/sync/errgroup
	go get github.com/go-playground/validator/v10
	go get github.com/google/uuid
	go get github.com/cenkalti/backoff/v4
	go get github.com/go-gomail/gomail
	go mod tidy
