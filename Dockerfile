# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache de modulos
COPY go.mod ./
RUN go mod download

# Codigo fonte
COPY . .

# Build estatico do binario
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/card-transaction ./cmd

FROM golang:1.22-alpine AS dev
WORKDIR /app

# Dependencias para desenvolvimento local no container
COPY go.mod ./
RUN go mod download

EXPOSE 8080
CMD ["go", "run", "./cmd/main.go"]

FROM alpine:3.20
WORKDIR /app

# Certificados para chamadas HTTPS (se necessario)
RUN apk --no-cache add ca-certificates

# Usuario nao-root para execucao segura da aplicacao
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/card-transaction /app/card-transaction
RUN chown app:app /app/card-transaction

EXPOSE 8080
USER app
ENTRYPOINT ["/app/card-transaction"]
