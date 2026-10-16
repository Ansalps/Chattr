# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# copy go mod first
COPY go.mod go.sum ./
RUN go mod download

# copy all files
COPY . .

# build binary
RUN go build -o post_relation_service ./cmd/main.go


# ---------- Stage 2: run ----------
FROM alpine:3.20

WORKDIR /app

# copy binary from builder
COPY --from=builder /app/post_relation_service .

# copy env file (important for viper)
#COPY pkg/config/dev.env ./pkg/config/dev.env

EXPOSE 50052

CMD ["./post_relation_service"]