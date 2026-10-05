# Build stage: static Go binary.
FROM golang:1.25-alpine AS go-builder

WORKDIR /build

# Module manifests first so dependency download is cached across source edits.
COPY go.mod go.sum ./
RUN go mod download

# build/ carries the //go:embed'd goose migrations; it must be copied with
# the sources or the embed fails to resolve.
COPY cmd ./cmd
COPY internal ./internal
COPY build ./build

# After sources, so a version bump does not bust the module cache layer.
ARG BUILD_VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X 'example.com/app/internal/pkg/buildutil.Version=${BUILD_VERSION}'" \
    -o /out/app ./cmd/app


# Runtime stage: distroless, non-root (uid 65532). One image serves every
# subcommand: `serve` (default), `migrate`, `seed`.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=go-builder /out/app /app

EXPOSE 3000

ENTRYPOINT ["/app"]
CMD ["serve"]
