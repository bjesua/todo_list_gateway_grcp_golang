# Dockerfile for gateway service
FROM golang:1.20 as builder

# Instalar buf directamente descargando el binario
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.45.0/buf-Linux-x86_64" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

# Instalar protoc-gen-go y protoc-gen-connect-go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
RUN go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@v1.0.0

WORKDIR /app

COPY . .

# Generar código gRPC con buf
RUN buf generate

RUN go mod tidy
RUN go build -o gateway .

FROM gcr.io/distroless/base-debian10

COPY --from=builder /app/gateway /app/gateway

CMD ["/app/gateway"]
