FROM golang:1.14.9-alpine AS builder
RUN mkdir /build
ADD go.mod go.sum main.go /build/
ADD controllers /build/controllers
WORKDIR /build
RUN go build

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/blockchain-playground /app/
WORKDIR /app
CMD ["./blockchain-playground"]