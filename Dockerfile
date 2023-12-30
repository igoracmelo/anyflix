FROM golang:1.21-alpine

WORKDIR /app

RUN apk add mpv

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build -v -o anyflix .

CMD ["./anyflix"]