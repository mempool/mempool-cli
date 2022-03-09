FROM golang:1.17-alpine@sha256:b35984144ec2c2dfd6200e112a9b8ecec4a8fd9eff0babaff330f1f82f14cb2a
RUN apk --no-cache add ca-certificates
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mempool-cli .

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /src/mempool-cli /
ENTRYPOINT ["/mempool-cli"]
