FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o storage-mcp-server .

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY --from=builder /app/storage-mcp-server /usr/local/bin/storage-mcp-server
ENTRYPOINT ["/usr/local/bin/storage-mcp-server"]
