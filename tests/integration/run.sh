#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $DIR

# starts the services (test-runner will do nothing here)
docker-compose up -d --build

# we manually run the test runner with its own start script
# this command will block until the test runner exits
docker-compose run test-runner /opt/scripts/start.sh

# we capture the return code of the test runner
rc=$?

# print result and clean up
echo -e "Completed with Exit code: ${rc}"
echo -e "Cleaning up docker..."
docker-compose down

# we exit with the same exit code as the test runner
# If succeeded the test will be marked as succeeded, otherwise it will be marked as failed
exit $rc
