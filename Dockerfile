FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.io,https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o async_account_check .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/async_account_check .
COPY --from=builder /app/etc ./etc
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
CMD ["./async_account_check", "-f", "./etc/config.yaml"]
