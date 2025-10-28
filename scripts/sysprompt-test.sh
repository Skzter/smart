#!/bin/sh

# in root
cd ..
echo "building ..."
go-task build

cd fixtures/autotester
fname="out/output.txt"
: > $fname 

for file in *; do 
    if [ -f "$file" ]; then 
        ./../../build/autotester &
        SERVER_PID=$!
        echo $SERVER_PID
        echo "Waiting for server to start..."

        # Wait until the server responds to curl
        # (Change URL as needed)
        URL="http://localhost:8081/v1/chat"
        MAX_RETRIES=20
        SLEEP_BETWEEN=1
        success=0

        for i in $(seq 1 $MAX_RETRIES); do
          if curl -s --max-time 2 "$URL" >/dev/null; then
            echo "Server is up! (attempt $i)"
            success=1
            break
          fi
          echo "Still waiting... ($i/$MAX_RETRIES)"
          sleep $SLEEP_BETWEEN
        done

        if [ $success -eq 0 ]; then
          echo "Server did not respond in time. Stopping..."
          kill $SERVER_PID
          exit 1
        fi

        # Now the server is running — do your test

        echo "Curling endpoint..."
        echo $file >> $fname
        curl "$URL" -d "$(cat $file)" | jq >> $fname 

        # When done, shut down the server
        echo "Shutting down server..."
        kill $SERVER_PID

        # Wait for it to stop
        wait $SERVER_PID 2>/dev/null
        sleep 2

        echo "Server stopped successfully."
    fi 
done
