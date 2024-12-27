FROM alpine:latest

ENV PORT=33333
RUN apk add busybox-extras

ENTRYPOINT telnet go-mud-server ${PORT}