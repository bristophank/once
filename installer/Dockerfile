# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o installer .

FROM scratch

COPY --from=builder /app/installer /installer
COPY --from=builder /app/templates /templates

EXPOSE 80

ENTRYPOINT ["/installer"]
