FROM golang:1.14-alpine
RUN apk --no-cache add ca-certificates
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mempool-cli .

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /src/mempool-cli /
ENTRYPOINT ["/mempool-cli"]
