FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY gen/ /app/gen/
COPY ctf-manager/go.mod ctf-manager/go.sum ./

RUN go mod download

COPY ctf-manager/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ctf-manager

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /ctf-manager .

EXPOSE 50052

CMD ["./ctf-manager"]
