# Use uma imagem base do Go
FROM golang:1.24-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git

# Definir diretório de trabalho
WORKDIR /app

# Copiar go mod e sum files
COPY go.mod go.sum ./

# Download das dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Imagem final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar o binário da aplicação
COPY --from=builder /app/main .

# Copiar pasta de documentos
COPY --from=builder /app/documents ./documents

# Expor porta
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
