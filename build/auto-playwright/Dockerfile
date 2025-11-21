FROM mcr.microsoft.com/playwright:v1.56.1-noble

RUN --mount=type=secret,id=git_credentials \
    GIT_TOKEN=$(cat /run/secrets/git_credentials) && \
    git clone https://gitlab-ci-token:${GIT_TOKEN}@gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/auto-playwright && \
    mv auto-playwright /app

WORKDIR /app

RUN npm install && npm run build
