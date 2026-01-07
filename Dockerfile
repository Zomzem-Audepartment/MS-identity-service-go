FROM golang:1.24-alpine AS builder

# Tăng tốc apk bằng cách không giữ cache
RUN apk add --no-cache git

WORKDIR /app

# --- TỐI ƯU 1: Cài tools riêng và tận dụng cache ---
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2

# --- TỐI ƯU 2: Cache Dependencies ---
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy toàn bộ code
COPY . .

# --- TỐI ƯU 3: Sqlc Generate ---
RUN sqlc generate

# --- TỐI ƯU 4: Build với BuildKit Cache ---
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binaries
COPY --from=builder /app/server /app/server
COPY --from=builder /go/bin/migrate /app/migrate

# Copy schema for migration
COPY --from=builder /app/sql/schema /app/sql/schema

# Copy entrypoint
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 4001

ENTRYPOINT ["/app/docker-entrypoint.sh"]
