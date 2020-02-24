FROM golang:1.13-alpine

WORKDIR /go/src/github.com/thatique/snowman
ADD . /go/src/github.com/thatique/snowman

RUN go install

FROM alpine:3.10

EXPOSE 6996 6997
RUN apk update \
    && apk add --no-cache --update ca-certificates openssl \
    && update-ca-certificates

COPY --from=0 /go/bin/snowman /usr/bin/snowman

CMD snowman