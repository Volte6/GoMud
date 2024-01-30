FROM alpine:latest

RUN apk add busybox-extras

COPY    provisioning/terminal/entrypoint.sh \
        Makefile \
        ./

RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
