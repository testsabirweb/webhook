

FROM golang:1.23-alpine as dev-env

WORKDIR /app

FROM dev-env as build-env
COPY go.mod /go.sum /app/
RUN go mod download

COPY . /app/

RUN CGO_ENABLED=0 go build -o /webhook

FROM alpine:3.10 as runtime

COPY --from=build-env /webhook /usr/local/bin/webhook
RUN chmod +x /usr/local/bin/webhook

ENTRYPOINT ["webhook"]