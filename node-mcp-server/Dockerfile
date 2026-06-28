# Build stage
FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o node-mcp-server .

# Final stage — minimal, no extra tools needed (no nsenter required here)
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY --from=builder /app/node-mcp-server /usr/local/bin/node-mcp-server
ENTRYPOINT ["/usr/local/bin/node-mcp-server"]
