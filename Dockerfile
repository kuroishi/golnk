# build stage
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o golnk ./cmd/golnk

# runtime stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/golnk .
RUN mkdir -p /data
VOLUME ["/data"]
ENV GOLNK_ADDR=":8080"
ENV GOLNK_FILE="/data/data.json"
EXPOSE 8080
CMD ["./golnk"]