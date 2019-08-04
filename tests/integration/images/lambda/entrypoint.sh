#!/usr/bin/env bash

# run the command passed as argument in the background
# (this is generally the lambda executable)
"$@" &

# give lambda 1 second to start
sleep 1

# Exposes LAMBDA gRPC port externally
# This is needed as LAMBDA listens on localhost
echo "* PROXYING 0.0.0.0:${LAMBDA_EXTERNAL_PORT} -> localhost:${_LAMBDA_SERVER_PORT}"
simpleproxy -L ${LAMBDA_EXTERNAL_PORT} -R localhost:${_LAMBDA_SERVER_PORT}
