FROM golang:1.24-alpine AS builder 
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN mkdir -p ./bin
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o ./bin/deadlinks .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/bin/deadlinks /usr/local/bin/deadlinks
ENTRYPOINT ["/usr/local/bin/deadlinks"]