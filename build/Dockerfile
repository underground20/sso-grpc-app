FROM golang:1.25.0-alpine3.21 AS deps

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

FROM deps AS builder

RUN apk --no-cache add \
    gcc \
    g++

COPY . .

RUN CGO_ENABLED=1 go build -o ./sso ./cmd/sso

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/sso /app/sso

CMD ["/app/sso"]