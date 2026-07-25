FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o ahb-panel .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ahb-panel .
EXPOSE 8080
CMD ["./ahb-panel"]
