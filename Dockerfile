FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/taifa-id \
    ./cmd/taifa-id

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/taifa-id-seed \
    ./cmd/taifa-id-seed

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S taifa \
    && adduser -S taifa -G taifa

COPY --from=build /out/taifa-id /usr/local/bin/taifa-id
COPY --from=build /out/taifa-id-seed /usr/local/bin/taifa-id-seed

USER taifa

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/taifa-id"]