# This dockerfile is specifically for running with tilt
FROM alpine
WORKDIR /
COPY ./scripts/manifests /manifests
COPY ./bin/manager /manager
ENTRYPOINT /manager
