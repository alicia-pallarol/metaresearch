# Multi-stage build. Build context is the backend/ directory, so nothing outside
# the Go module is ever copied into an image layer.
#
#   docker build -t atlas-api ./backend
#   docker run -p 8080:8080 --env-file backend/.env atlas-api

FROM golang:1.24-alpine AS build

WORKDIR /src

# Dependencies first so an unchanged go.mod/go.sum keeps this layer cached.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off gives a static binary; trimpath and -w -s keep it small and free of
# local build paths.
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/server .

# Distroless: no shell, no package manager, non-root by default.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

# Render injects PORT; 8080 is the default the binary falls back to.
EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/server"]
