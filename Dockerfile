FROM golang:1.27.2-alpine AS builder

WORKDIR /build

COPY go.mod ./

COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -trimpath -o app ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder /build/app .

EXPOSE 2347

USER nonroot

ENTRYPOINT ["/app/app"]

LABEL org.opencontainers.image.title="netkit" \
    org.opencontainers.image.description="Network and developer toolkit: IP, DNS, redirects, SSL and JSON tools" \
    org.opencontainers.image.source="https://github.com/enteresanlikk/netkit"
