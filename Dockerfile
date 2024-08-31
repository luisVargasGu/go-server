FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o app .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
