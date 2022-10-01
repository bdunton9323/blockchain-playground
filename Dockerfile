FROM golang:1.19.1-alpine AS builder
RUN apk --update --no-cache add g++
RUN mkdir /build
ADD go.mod go.sum main.go /build/
ADD controllers /build/controllers
ADD contract /build/contract
WORKDIR /build
RUN go build

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/blockchain-playground /app/
WORKDIR /app
CMD ["./blockchain-playground"]