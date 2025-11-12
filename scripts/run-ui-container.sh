#!/bin/sh

FILE="${1:-mock.spec.ts}"
BASEFILE=$(basename $FILE)
cd ..

echo "starting docker container"

docker run --rm \
    --env OPENAI_API_KEY="$(doppler secrets get OPENAI_KEY --plain)" \
    -v "$PWD/$FILE":/app/$BASEFILE \
    -v "$PWD/docker/logs":/app/logs/ \
    --network=host \
    auto-pw:latest \
    /bin/bash -c "cd /app && npx playwright test $BASEFILE --ui-port=3000 --ui-host=0.0.0.0"

echo "docker container finished"

echo "cleaning logs of color codes"

sed -i "s/\x1B\[[0-9;]*[mGKHF]//g" docker/logs/output.log

echo "cleaned logfile in /docker/logs/output.log"
