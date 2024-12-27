FROM golang:1.23.4-alpine3.21 AS builder

ARG BIN=go-mud-server
ENV BIN=${BIN}
WORKDIR /src

RUN apk add --update make bash

# Create logfile
RUN mkdir -p /app/log
RUN touch /app/log/logfile.log

COPY . .
RUN go build -o ${BIN}


FROM alpine:latest AS runner
ARG BIN=go-mud-server
ENV BIN=${BIN}
WORKDIR /app

EXPOSE ${PORT}
COPY --from=builder /src/${BIN} .
COPY --from=builder /src/_datafiles ./_datafiles

ENTRYPOINT ["/bin/sh", "-c", "./${BIN}"]