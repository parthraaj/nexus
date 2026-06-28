# Build stage
FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o node-daemon .

# Final stage — minimal image with nsenter available
# util-linux provides nsenter
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
RUN microdnf install -y util-linux && microdnf clean all
COPY --from=builder /app/node-daemon /usr/local/bin/node-daemon
ENTRYPOINT ["/usr/local/bin/node-daemon"]
