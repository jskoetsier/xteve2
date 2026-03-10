# Stage 1: Build React frontend
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./internal/ui/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o xteve ./cmd/xteve/

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ffmpeg ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/xteve .
EXPOSE 34400
VOLUME ["/config"]
ENTRYPOINT ["/app/xteve", "-config", "/config"]
