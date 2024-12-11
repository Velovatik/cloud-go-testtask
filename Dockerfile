FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go mod tidy
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

# Reference: https://github.com/coding-latte/golang-docker-multistage-build-demo/blob/master/alpine.dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/migrate .

# Add goose CLI
COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY ./migrations ./migrations
COPY ./config ./config

EXPOSE 8082

CMD ["./app"]