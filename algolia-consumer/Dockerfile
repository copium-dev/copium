# build stage
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go

# final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
# for algolia api keys
COPY --from=builder /app/.env .
# in prod we dont use pubsub emulator
ENV ENVIRONMENT=prod    
EXPOSE 8080
CMD ["./main"]