# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# copy mod files
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# build server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot_server ./server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# copy bin
COPY --from=builder /app/bot_server .

# copy notifier.config
COPY --from=builder /app/notifier.config .

EXPOSE 8081

# run bot
CMD ["./bot_server"]
