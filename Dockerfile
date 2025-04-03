# Build stage
FROM golang:1.24.2 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o main .

# Final stage
FROM gcr.io/distroless/base-debian12
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["/root/main"]
