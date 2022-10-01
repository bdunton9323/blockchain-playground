FROM golang:1.14.9-alpine AS builder
RUN mkdir /build
ADD go.mod go.sum hello.go /build/
WORKDIR /build
RUN go build

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
# Note to self: "go build" compiles it to a file matching the module name in go.mod
COPY --from=builder /build/blockchain-playground /app/
WORKDIR /app
CMD ["./blockchain-playground"]