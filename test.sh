#!/bin/bash

set -euo pipefail

RUNALL=false
DEPENDENCIES=false
TESTS=false
BUILD=false

print_usage() {
  echo "Program usage: [./script.sh flags]

        Flag: -a [Run everything bellow]
        Flag: -d [Get dependencies]
        Flag: -t [Run tests with coverage]
        Flag: -b [Run a build as test]

        Flag: -h [Print help message]
        "
}

while getopts 'hadtb' flag; do
  case "${flag}" in
    h) print_usage
        exit 0 ;;
    a) RUNALL=true ;;
    d) DEPENDENCIES=true ;;
    t) TESTS=true ;;
    b) BUILD=true ;;
    *) print_usage
       exit 1 ;;
  esac
done

if [[ ${DEPENDENCIES} == true || ${RUNALL} == true ]]; then
  echo "Get dependencies"
  go get -v -t -d ./...
fi

if [[ ${TESTS} == true || ${RUNALL} == true ]]; then
  echo "Run test with coverage"
  go test -v ./... -race -coverprofile=coverage.txt -covermode=atomic
fi

if [[ ${BUILD} == true || ${RUNALL} == true ]]; then
  echo "Run a build as test"
  go build -v cmd/lcm/main.go
fi