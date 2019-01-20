#!/bin/bash
for os in "darwin" "linux" "windows"; do
    for arch in "amd64" "386"; do
        GOOS=$os GOARCH=$arch go build -o ./bin/govtvfetch-$os-$arch ./cmd/govtvfetch
    done
done
