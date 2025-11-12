FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /decision-service ./...

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /decision-service /usr/local/bin/decision-service
EXPOSE 8080
USER nobody
ENTRYPOINT ["/usr/local/bin/decision-service"]
