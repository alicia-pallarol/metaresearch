# Multi-stage build for the Go backend. Build context is the repo root so the
# image can bundle the /migrations directory alongside the binary.
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /src/backend
RUN go mod download

# Build.
WORKDIR /src
COPY backend ./backend
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server /app/server
# Migrations are read at startup; ship them next to the binary.
COPY migrations /app/migrations
ENV MIGRATIONS_PATH=/app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
