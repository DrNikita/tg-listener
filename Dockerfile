FROM golang:1.23.4 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o tg-listener

FROM golang:1.23.4
COPY --from=builder /app/tg-listener /
COPY wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/tg-listener"]