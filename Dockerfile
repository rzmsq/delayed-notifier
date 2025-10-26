# Builder stage
FROM golang:latest AS builder

WORKDIR /app

# Copy dependency files and download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the server and consumer binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/email_consumer ./cmd/consumer/email
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/telegram_consumer ./cmd/consumer/telegram

# Final stage
FROM alpine:3.20

# Install certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the configuration file from the source context
COPY ./config.yaml .

# Copy the compiled binaries from the builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/email_consumer .
COPY --from=builder /app/telegram_consumer .

# Set the command to run the server
CMD ["/app/server"]
