#!/bin/bash

mkdir dist
export GOARCH=amd64
export GOOS=darwin && go build -ldflags="-s -w -X main.Version=$TRAVIS_TAG" -o dist/lcm-$TRAVIS_TAG-darwin cmd/lcm/main.go
export GOOS=linux && go build -ldflags="-s -w -X cmd.lcm.main.Version=$TRAVIS_TAG" -o dist/lcm-$TRAVIS_TAG-linux cmd/lcm/main.go
export GOOS=windows && go build -ldflags="-s -w -X cmd.lcm.main.Version=$TRAVIS_TAG" -o dist/lcm-$TRAVIS_TAG-windows cmd/lcm/main.go

ls -la dist
