FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/w2g ./cmd/w2g

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /out/w2g /app/w2g
COPY web /app/web

EXPOSE 8080

ENV PORT=8080
ENV STORAGE_DIR=/app/storage

CMD ["/app/w2g"]
