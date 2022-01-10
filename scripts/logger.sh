#!/bin/bash


cd $(dirname "$0")

# TODO: validate files and folders [botond.pelyi]

docker run \
  -v ${PWD}/../db:/app/db:rw \
  -v ${PWD}/../bin/logger:/app/logger:ro \
  -v ${PWD}/../configs/config.dev.yaml:/app/config.yaml:ro \
  alpine:3.15.0 \
  ash -c "CONFIG_FILE=/app/config.yaml /app/logger"
