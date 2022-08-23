# quic-chat

A simple chat that works over [QUIC](https://en.wikipedia.org/wiki/QUIC).

## Running Server

1. Generate a set of private and public keys:

```shell
openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

2. Run the server:

```shell
go run ./cmd/server
```

## Running Client

Run the client:

```shell
go run ./cmd/clinet [-s <ServerAddress>] -n <Nickname>
```
