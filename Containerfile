FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pod-lens .

FROM alpine:3.19

ARG THEME=light

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

ENV THEME=${THEME}

COPY --from=builder /app/pod-lens .

COPY --chown=app:app pages/ ./pages/
COPY --chown=app:app static/ ./static/

USER app

ENV PORT=80

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 80 || exit 1

# Labels
LABEL org.opencontainers.image.title="Pod Lens" \
      org.opencontainers.image.description="Lightweight pod information viewer similar to traefik/whoami" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.source="https://github.com/williampsena/pod-lens"

CMD ["./pod-lens"]
