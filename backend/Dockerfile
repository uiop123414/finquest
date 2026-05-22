FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && go build -o finquest ./main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/finquest .
COPY --from=builder /app/db/migrations ./db/migrations
EXPOSE 8000
CMD ["./finquest"]
