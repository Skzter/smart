#!/bin/sh

echo "Going to dockerfile"

cd ../docker 

echo "Building Image"

docker build -t auto-pw .
