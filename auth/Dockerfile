FROM golang:1.21 as builder

# Instalar buf directamente descargando el binario
RUN curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.45.0/buf-Linux-x86_64" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

# Instalar protoc-gen-go y protoc-gen-connect-go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest

WORKDIR /app

COPY . .

# Generar código gRPC con buf
RUN buf generate

# Ejecutar go mod tidy y construir el binario estáticamente para ARM64
RUN go mod tidy
RUN CGO_ENABLED=0 GOARCH=arm64 go build -o /app/auth-binary .

# Usa una imagen mínima porque no necesitamos GLIBC
FROM scratch

# Copiar el binario correctamente
COPY --from=builder /app/auth-binary /app/auth-binary

# Ejecutar el binario directamente sin depender de GLIBC
CMD ["/app/auth-binary"]
