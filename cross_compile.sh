#!/usr/bin/env bash

# supposes that you have the following package installed:
# pacman -S extra/mingw-w64-gcc extra/mingw-w64-headers
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build .

