#!/bin/sh

set -e -x

# Build toml-test.
env "GOBIN=$PWD" go install github.com/BurntSushi/toml-test/cmd/toml-test@master

# Build test adapter.
go build -o toml-test-adapter toml-test-main.go

# Run decoder tests.
./toml-test ./toml-test-adapter

# Run encoder tests.
./toml-test -encoder -- ./toml-test-adapter -e
