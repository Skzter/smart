#!/bin/sh

FILE="$1"

# Check if an argument was provided
if [ -z "$FILE" ]; then
    echo "Error: No file specified"
    echo "Usage: $0 <file>"
    exit 1
fi

# Check if the file exists
if [ ! -e "$FILE" ]; then
    echo "Error: File '$FILE' does not exist"
    exit 1
fi

BASEFILE=$(basename $FILE)
LOGDIR="logs/"

# logs/uuid.spec.ts.log
LOGPATH=$LOGDIR$BASEFILE.log 

cd ..

echo "starting docker container"

docker run --rm \
    --env OPENAI_API_KEY="$(doppler secrets get OPENAI_KEY --plain)" \
    -v "$FILE":/app/$BASEFILE \
    -v "$PWD/$LOGDIR":/app/$LOGDIR \
    --name="auto-playwright-headless" \
    --network=host \
    auto-pw:latest \
    /bin/bash -c "cd /app && npx playwright test $BASEFILE --reporter=list > $LOGPATH 2>&1"

echo "docker container finished"

echo "cleaning logs of color codes"

sed -i 's/\x1B\[[0-9;]*[mGKHF]//g' $LOGPATH

echo "cleaned logfile in $LOGPATH"
