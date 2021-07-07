#!/bin/bash
export CGO_ENABLED=0
go build -ldflags "-s -w" -o ../../bin/echo-server .
