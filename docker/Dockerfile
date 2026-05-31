FROM golang:1.22-alpine AS dev
ARG APP
ENV APP=$APP
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["sh", "-c", "air -c .air.${APP}.toml"]

FROM golang:1.17.5-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG APP

ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN apk add --no-cache file git

WORKDIR /src

COPY . /src

RUN go build -o /out/wl-app /src/cmd/$APP

FROM alpine:3.15.0 AS app

COPY --from=builder /out/wl-app /usr/local/bin/wl-app

ENTRYPOINT ["wl-app"]
