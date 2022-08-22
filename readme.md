# quic-chat

## Run Server

1. Navigate to the `tls` directory and generate a set of private and public keys:

```shell
cd tls
openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```