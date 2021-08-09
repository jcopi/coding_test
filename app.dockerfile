FROM golang:1.16.7-alpine3.14 as golang

WORKDIR /go/src/github.com/jcopi/coding_test
COPY main.go .
COPY go.mod .

RUN apk add --update git gcc
RUN go get github.com/gin-gonic/gin
RUN go get github.com/google/uuid
RUN go get go.uber.org/zap
RUN go get go.etcd.io/etcd/client/v3

RUN go build -o app

FROM alpine:3.14

ARG USER=api
ENV HOME=/api
RUN mkdir $HOME
RUN adduser -D $USER \
    && chown $USER $HOME \ 
    && chmod +x $HOME

USER $USER
WORKDIR $HOME

COPY --from=golang "/go/src/github.com/jcopi/coding_test/app" .

EXPOSE 8000
ENTRYPOINT ["./app"]