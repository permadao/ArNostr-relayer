FROM alpine:latest
MAINTAINER sandy <sandy@ever.finance>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

WORKDIR /relayer

COPY basic/relayer /relayer/relayer
EXPOSE 8080

ENTRYPOINT [ "/relayer/relayer" ]