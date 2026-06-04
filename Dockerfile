# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/OpenSustain ./cmd/OpenSustain

# ── Stage 2: Minimal runtime ──────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/OpenSustain /OpenSustain

# GitHub Actions injects args from action.yml; ENTRYPOINT sets the binary
ENTRYPOINT ["/OpenSustain"]
