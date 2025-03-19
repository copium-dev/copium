# build stage
FROM python:3.10-slim AS builder
WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY . .

EXPOSE 8080
CMD ["python", "main.py"]