FROM golang:1.12-alpine

ENV GOPATH /go
ENV CGO_ENABLED 0
ENV GO111MODULE on

WORKDIR /go/src/github.com/thatique/snowman
ADD . /go/src/github.com/thatique/snowman

RUN \
    apk add --no-cache git && \
    go install

FROM alpine:3.9

EXPOSE 6996 6997
RUN apk update \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates \

COPY --from=0 /go/bin/snowman /usr/bin/snowman

CMD snowman