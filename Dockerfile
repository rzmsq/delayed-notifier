FROM golang:latest

WORKDIR /delayed-notifier

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /delayed-notifier/server ./cmd/server

EXPOSE 8000

CMD ["/delayed-notifier/server"]