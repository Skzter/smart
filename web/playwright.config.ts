// playwright.config.ts
import { defineConfig } from "@playwright/test";

export default defineConfig({
    testDir: "./tests/e2e", // your E2E test directory
    use: {
        baseURL: "http://localhost:5173",
        headless: true,
    },
    webServer: {
        command: "pnpm dev",
        port: 5173,
        reuseExistingServer: !process.env.CI, // don't restart if already running
    },
});
