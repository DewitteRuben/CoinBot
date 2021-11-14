FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod /app

COPY go.sum /app


RUN go mod download

COPY . /app

RUN export CGO_ENABLED=0

RUN go build -o bot

ENTRYPOINT ["sh", "./init.sh"]
