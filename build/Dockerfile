# syntax=docker/dockerfile:1

# ---- 构建阶段 ----
FROM golang:1.25 AS builder

ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

WORKDIR /src

# 先拷贝依赖清单，利用层缓存
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译静态二进制
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
    -o /out/apiserver ./cmd/apiserver

# ---- 运行阶段 ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/apiserver /app/apiserver
COPY configs/config.yaml /app/configs/config.yaml

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/apiserver"]
CMD ["-c", "/app/configs/config.yaml"]
