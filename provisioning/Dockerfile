FROM golang:1.21.3-alpine3.18 as builder

ARG BIN=app
ENV BIN ${BIN}

RUN apk add --update make bash
RUN mkdir -p /app
WORKDIR /src

COPY    . ./

RUN make build_only
RUN cp -R   provisioning/entrypoint.sh \
            _datafiles \
            ${BIN} \
            /app/

FROM alpine:latest

ARG BIN=app
ENV BIN ${BIN}

EXPOSE ${PORT}

WORKDIR /app
COPY --from=builder /app .

RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
