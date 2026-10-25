# ==================== Build Stage ====================
FROM golang:1.26-alpine AS builder

WORKDIR /build

# 依赖层单独缓存：go.mod / go.sum 未变时不会重复下载模块
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 数据库驱动为纯 Go 实现（modernc.org/sqlite），不依赖 CGO / gcc，
# 产物是静态二进制，可直接跑在无编译器的精简镜像上。
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /build/bin/server ./cmd/api

# ==================== Runtime Stage ====================
FROM alpine:3.20

# tzdata 必需：权限记录的 expire_at 是绝对时间，容器缺时区库会导致时间偏移
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /build/bin/server /app/server
COPY config.example.yaml /app/config.yaml

# SQLite 数据目录与日志目录需要写权限（默认 dsn 为 data/landc_permission.db）
RUN mkdir -p /app/data /app/logs && chown -R app:app /app

USER app

ENV LANDC_CONFIG=/app/config.yaml \
    TZ=Asia/Shanghai

EXPOSE 8082

# 框架内置存活性端点（server.health_check.enabled 默认开启）
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8082/health || exit 1

ENTRYPOINT ["/app/server"]
