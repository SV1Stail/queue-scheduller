FROM golang:1.24.2 AS builder

WORKDIR /app

ADD ../ ./

RUN go mod download

RUN go build -o /bin/queue-scheduler /app/main.go

ENTRYPOINT ["/bin/queue-scheduler"]
