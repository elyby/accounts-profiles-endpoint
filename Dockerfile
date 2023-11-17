# syntax=docker/dockerfile:1

FROM golang:1.21 AS builder

COPY . /build
WORKDIR /build
RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build \
    -trimpath \
    -ldflags="-w -s" \
    -o app \
    main.go

FROM scratch

COPY --from=builder /build/app /root/app

ENTRYPOINT ["/root/app"]
EXPOSE 8080
