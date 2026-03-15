FROM golang:1.25-alpine
WORKDIR /app

COPY . .
RUN go build -o app ./cmd/bot

CMD ["./app"]