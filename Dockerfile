# Build a fully static binary (modernc SQLite is pure Go — no cgo).
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Generated *_templ.go files are committed, so no templ step is needed here.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/aniicrite ./cmd/server

# Minimal runtime image.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/aniicrite /app/aniicrite
ENV APP_ADDR=":8080" DATA_DIR="/data"
VOLUME ["/data"]
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/app/aniicrite"]
