FROM mcr.microsoft.com/playwright:v1.56.1-noble

RUN git clone https://github.com/lucgagan/auto-playwright.git && \
    mv auto-playwright /app

WORKDIR /app

RUN sed -i 's/"@playwright\/test": "1\.52\.0"/"@playwright\/test": "1.56.1"/g' package.json

RUN npm install && npm run build
