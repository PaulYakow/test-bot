# syntax=docker/dockerfile:1

FROM golang:1.21-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o app ./cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 80
CMD ["/app/app"]