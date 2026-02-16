# ============================================================
# GitOps-Time-Machine — Multi-Stage Docker Build
# ============================================================

# Stage 1: Build
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /gitops-time-machine .

# Stage 2: Runtime (minimal)
FROM alpine:3.19

RUN apk add --no-cache ca-certificates git

COPY --from=builder /gitops-time-machine /usr/local/bin/gitops-time-machine

# Create non-root user
RUN adduser -D -h /home/gtm gtm
USER gtm
WORKDIR /home/gtm

ENTRYPOINT ["gitops-time-machine"]
CMD ["--help"]
