# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/rest ./cmd/rest

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app

COPY --from=builder /app/bin/rest .

EXPOSE 8080
ENTRYPOINT ["./rest"]
