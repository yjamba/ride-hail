FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git build-base ca-certificates

WORKDIR /src

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Build args
ARG BUILD_TARGET=./cmd/auth
ARG CGO_ENABLED=0

# Copy sources and build
COPY . .
RUN CGO_ENABLED=$CGO_ENABLED GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/main $BUILD_TARGET

FROM alpine:3.18
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/main /app/main

EXPOSE 3001
ENV PORT=3001

CMD ["/app/main"]
