FROM golang:1.22 AS builder

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o ./app github.com/Alp4ka/classifier-telegram/cmd/app

FROM alpine:latest AS run

RUN apk add curl

COPY --from=builder /build/app /usr/local/bin/app

CMD ["app"]
