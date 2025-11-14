FROM mcr.microsoft.com/playwright:v1.56.1-noble

# forked image for improved usage
RUN git clone https://gitlab.dit.htwk-leipzig.de/bruno.bittner/auto-playwright && \
    mv auto-playwright /app

WORKDIR /app

RUN npm install && npm run build
