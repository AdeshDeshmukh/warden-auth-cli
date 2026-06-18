FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=local
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -ldflags="-w -s" \
    -o warden ./cmd/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/warden .

ENTRYPOINT ["/app/warden"]