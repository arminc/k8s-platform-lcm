#!/bin/bash

set -euo pipefail

RUNALL=false
DEPENDENCIES=false
TESTS=false
BUILD=false
LINTER=false

print_usage() {
  echo "Program usage: [./script.sh flags]

        Flag: -a [Run everything bellow]
        Flag: -d [Get dependencies]
        Flag: -t [Run tests with coverage]
        Flag: -b [Run a build as test]
        Flag: -l [Run golangci-lint]

        Flag: -h [Print help message]
        "
}

while getopts 'hadtbl' flag; do
  case "${flag}" in
    h) print_usage
        exit 0 ;;
    a) RUNALL=true ;;
    d) DEPENDENCIES=true ;;
    t) TESTS=true ;;
    b) BUILD=true ;;
    l) LINTER=true ;;
    *) print_usage
       exit 1 ;;
  esac
done

if [[ ${DEPENDENCIES} == true || ${RUNALL} == true ]]; then
  echo "Get dependencies"
  go get -v -t -d ./...
fi

if [[ ${LINTER} == true || ${RUNALL} == true ]]; then
  echo "Run golangci-lint"
  golangci-lint run
fi

if [[ ${TESTS} == true || ${RUNALL} == true ]]; then
  echo "Run test with coverage"
  go test -v ./... -race -coverprofile=coverage.txt -covermode=atomic
fi

if [[ ${BUILD} == true || ${RUNALL} == true ]]; then
  echo "Run a build as test"
  go build -v cmd/lcm/main.go
fi