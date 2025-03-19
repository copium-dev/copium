# build stage
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/jtracker-backend-credentials.json .
COPY --from=builder /app/pubsub-credentials.json .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]