# image for building
FROM golang:1.25-alpine AS builder
WORKDIR /app

# copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY *.go ./
COPY ai/ ./ai/

# build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o decision_service .


# create final light weight image
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# copy the binary from builder
COPY --from=builder /app/decision_service .

# expose port
EXPOSE 8081

# run the binary
CMD ["./decision_service"]