FROM golang:1.21 AS builder
ENV CGO_ENABLED 0
ADD . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -v -o universal-store-api .

FROM alpine:3
RUN apk update && \
    apk add openssl && \
    rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

ADD Dockerfile /Dockerfile

COPY --from=builder /app/universal-store-api /app/universal-store-api

RUN chown nobody /app/universal-store-api \
    && chmod 500 /app/universal-store-api

USER nobody

ENTRYPOINT ["/app/universal-store-api"]
