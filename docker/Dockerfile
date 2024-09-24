FROM golang:1.17.5-alpine as builder

WORKDIR /source

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/$APP ./cmd/$APP

FROM alpine:3.15.0

COPY --from=builder /out/$APP /usr/local/bin/$APP

CMD ["$APP"]
