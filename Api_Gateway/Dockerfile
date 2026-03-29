# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# copy go mod first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# copy all files
COPY . .

# build binary
RUN go build -o api_gateway ./cmd/main.go


# ---------- Stage 2: run ----------
FROM alpine:3.20

WORKDIR /app

# copy binary from builder
COPY --from=builder /app/api_gateway .

# ✅ keep config path same
COPY pkg/config/dev.env ./pkg/config/dev.env

# ✅ keep template path same
COPY cmd/templates ./cmd/templates

EXPOSE 3000

CMD ["./api_gateway"]