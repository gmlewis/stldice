#!/bin/bash -ex
# -*- compile-command: "./build-push.sh"; -*-
CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' ./stl2svx-server.go
# ldd ./stl2svx-server
docker build -t us.gcr.io/gmlewis/stl2svx-server .
gcloud docker -- push us.gcr.io/gmlewis/stl2svx-server:latest
