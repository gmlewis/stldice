FROM alpine:latest

# Add trusted CA root bundles
RUN   apk update \
  &&   apk add ca-certificates wget \
  &&   update-ca-certificates

VOLUME /data
WORKDIR "/"

COPY ./stl2svx-server /

ENTRYPOINT ["/stl2svx-server"]
