# Build stage
FROM golang:1.25-bookworm AS builder
WORKDIR /workspace

# Cache modules
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download

# Copy source
COPY . .

# Build static binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /workspace/server ./

# Final stage - small runtime image with CA certs
FROM gcr.io/distroless/cc-debian11
WORKDIR /
COPY --from=builder /workspace/server /server

# Cloud Run expects the server to listen on $PORT (default 8080)
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/server"]
