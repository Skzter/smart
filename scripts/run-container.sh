#!/bin/sh

FILE="${1:-test.spec.ts}"
cd ..

echo "starting docker container"

docker run --rm \
    --env OPENAI_API_KEY="$(doppler secrets get OPENAI_KEY --plain)" \
    -v "$PWD/fixtures/auto-playwright/$FILE":/app/$FILE \
    -v "$PWD/docker/logs":/app/logs/ \
    auto-pw:latest \
    /bin/bash -c "cd /app && npx playwright test test.spec.ts --reporter=list > logs/output.log 2>&1"

echo "docker container finished"

echo "cleaning logs of color codes"

sed -i "s/\x1B\[[0-9;]*[mGKHF]//g" docker/logs/output.log

echo "cleaned logfile in /docker/logs/output.log"
