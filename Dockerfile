FROM golang:1.27.3-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gomark .

FROM alpine:3.22

RUN addgroup -S gomark && adduser -S -G gomark -u 10001 gomark

WORKDIR /app

COPY --from=build /out/gomark /usr/local/bin/gomark
COPY --chown=gomark:gomark posts ./posts

ENV BLOG_ADDR=:8080 \
    BLOG_POSTS=/app/posts

USER gomark

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["gomark"]
