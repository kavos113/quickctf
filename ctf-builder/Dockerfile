FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY gen/ /gen/
COPY lib/ /lib/
COPY ctf-builder/go.mod ctf-builder/go.sum ./

RUN go mod download

COPY ctf-builder/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ctf-builder

FROM docker:dind

COPY --from=builder /ctf-builder /usr/local/bin/

EXPOSE 50051

CMD ["ctf-builder"]
