#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $DIR

docker-compose up -d --build
docker-compose run test-runner /opt/scripts/start.sh

rc=$?
echo -e "Completed with Exit code: ${rc}"

echo -e "Cleaning up docker..."
docker-compose down

exit $rc
