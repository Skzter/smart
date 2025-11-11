#!/bin/sh

# in root dir of project
cd ..

docker run -it --env OPENAI_API_KEY="$(doppler secrets get OPENAI_KEY --plain)" -v "$PWD/fixtures/auto-playwright/test.spec.ts":/app/test.spec.ts auto-pw:latest /bin/bash
