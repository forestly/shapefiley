FROM golang

MAINTAINER Abhi Yerra <abhi@berkeley.edu>

ADD . /go/src/github.com/forestly/shapefiley

RUN cd /go/src/github.com/forestly/shapefiley && go get ./...
RUN go install github.com/forestly/shapefiley

WORKDIR /go/src/github.com/forestly/shapefiley

ENTRYPOINT /go/bin/shapefiley

EXPOSE 3002
