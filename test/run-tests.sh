#!/bin/sh
set -eu

go mod download

exec make testacc-local TEST=./internal/provider/...
