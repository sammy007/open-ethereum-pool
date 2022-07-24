#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

GOBIN="$PWD/build/bin"
export GOBIN

exec "$@"
