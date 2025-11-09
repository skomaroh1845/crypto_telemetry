# image for building
FROM golang:1.23-alpine AS builder
WORKDIR /app

# copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY *.go ./

# build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o data_service .


# create final light weight image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# copy the binary from builder
COPY --from=builder /app/data_service .

# expose port
EXPOSE 8080

# run the binary
CMD ["./data_service"]