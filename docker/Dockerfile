FROM golang:1.17.5-alpine as builder

ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN apk add --no-cache file git

WORKDIR /src

COPY . /src

RUN printenv

RUN go build -o /out/$APP /src/cmd/$APP

FROM alpine:3.15.0 as app

COPY --from=builder /out/$APP /usr/local/bin/$APP

CMD ["$APP"]
