FROM golang:1.20.3-alpine3.16 as builder

WORKDIR /app
COPY . /app
RUN mkdir -p build
RUN go build -o build/msgpack main.go

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/build/msgpack /app
CMD ["./msgpack"]