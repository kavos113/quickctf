FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY gen/ /app/gen/
COPY ctf-runner/go.mod ctf-runner/go.sum ./

RUN go mod download

COPY ctf-runner/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ctf-runner

FROM docker:dind

COPY --from=builder /ctf-runner /usr/local/bin/

EXPOSE 50053

CMD ["ctf-runner"]
