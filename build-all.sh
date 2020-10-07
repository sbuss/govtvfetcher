#!/bin/bash
for os in "darwin" "linux" "windows"; do
    for arch in "amd64" "386"; do
        out="govtvfetch-$os-$arch"
        if [[ $os == "windows" ]]; then
            out="$out.exe"
        fi
        GOOS=$os GOARCH=$arch go build -o ./bin/$out ./cmd/govtvfetch
    done
done
