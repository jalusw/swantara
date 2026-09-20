FROM golang:1.25-alpine AS builder
WORKDIR /src/service

COPY service/go.mod service/go.sum ./
RUN go mod download

COPY service/. .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/http ./cmd/http/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/queue ./cmd/queue/main.go
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.3

FROM alpine:3.20 AS runner
WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder --chown=appuser:appgroup /out/http ./http
COPY --from=builder --chown=appuser:appgroup /out/queue ./queue
COPY --from=builder --chown=appuser:appgroup /go/bin/goose ./goose
COPY --from=builder --chown=appuser:appgroup /src/service/migrations ./migrations

USER appuser

EXPOSE 8080

CMD ["./http"]