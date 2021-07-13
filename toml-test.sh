#!/bin/sh

# This script builds and runs the toml-test suite.
#
# Running all decoder tests:
#
#   ./toml-test.sh
#
# Or encoder tests:
#
#   ./toml-test.sh -encoder
#
# Or a specific encoder test:
#
#   ./toml-test.sh -run valid/string/escapes -encoder

set -e

# Separate arguments.
args=""
testargs=""
while [ $# -gt 0 ] ; do
    case "$1" in
        "-encoder")
            args="$args $1"
            testargs="$testargs -e" ;;
        "--")
            shift
            testargs="$testargs $@"
            break ;;
        *)
            args="$args $1" ;;
    esac
    shift
done

set -x

# Build toml-test.
version=2349618fe2bcc4393461c6c7d37b417c05e1b181
env "GOBIN=$PWD" go install "github.com/BurntSushi/toml-test/cmd/toml-test@$version"

# Build test adapter.
go build -o toml-test-adapter toml-test-adapter.go

# Run the tests.
./toml-test $args -- ./toml-test-adapter $testargs
