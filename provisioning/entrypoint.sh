#!/bin/sh

if [[ -z "$1" && -n "${BIN:=}" ]]; then
    set -- ./${BIN}
fi

exec "$@"

