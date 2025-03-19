# build stage; use debian for better compatibility
FROM golang:bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main main.go

# final stage; can switch to alpine
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
# in prod we dont use pubsub emulator
ENV ENVIRONMENT=prod
EXPOSE 8080
CMD ["./main"]