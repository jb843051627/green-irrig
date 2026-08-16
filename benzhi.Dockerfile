# benzhi.Dockerfile — 保留完整 Go 工具链，支持 arm64+amd64
FROM golang:1.22-bookworm

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "./..."]
