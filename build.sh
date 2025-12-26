#!/bin/bash
CGO_ENABLED=0 go build -o b64 --ldflags="-s -w" --trimpath