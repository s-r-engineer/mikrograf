FROM golang:1.24.2 AS builder

COPY . /build/

WORKDIR /build

RUN go build -trimpath -o mikrograf .

FROM telegraf:1.32.3

COPY --from=builder /build/mikrograf /usr/local/bin/mikrograf