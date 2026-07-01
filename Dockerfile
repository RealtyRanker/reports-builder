FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /reports-builder ./cmd/reportsbuilder

# ---

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN mkdir -p /var/log/reports-builder

WORKDIR /app
COPY --from=builder /reports-builder .
COPY config.yaml .

EXPOSE 9096

CMD ["./reports-builder", "-config", "config.yaml"]
