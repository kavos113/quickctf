FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY gen/ /gen/
COPY lib/ /lib/
COPY ctf-server/go.mod ctf-server/go.sum ./

RUN go mod download

COPY ctf-server/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ctf-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /ctf-server .
COPY migration/ /root/migration/

EXPOSE 50060

CMD ["./ctf-server"]
