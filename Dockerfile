# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder

ARG VERSION=undefined
ARG COMMIT=unknown

COPY . /build
WORKDIR /build
RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build \
    -trimpath \
    -ldflags="-w -s -X ely.by/sessionserver/internal/version.version=$VERSION -X ely.by/sessionserver/internal/version.commit=$COMMIT" \
    -o app \
    main.go

FROM scratch

COPY --from=builder /build/app /root/app

ENTRYPOINT ["/root/app"]
EXPOSE 8080
