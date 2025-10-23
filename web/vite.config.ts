/// <reference types="vitest/config" />
import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import { svelteTesting } from "@testing-library/svelte/vite";

// https://vite.dev/config/
export default defineConfig({
    plugins: [tailwindcss(), svelte(), svelteTesting()],
    test: {
        globals: true,
        environment: "jsdom",
        setupFiles: "./src/setupTests.ts",
        include: ["tests/unit/*.test.ts", "tests/components/*.test.ts"],
        reporters: ["junit", "default"],
        outputFile: "coverage/junit.xml",
        coverage: {
            provider: "v8",
            reporter: ["text", "cobertura", "lcov"],
            include: ["src/{lib,components}/**/*.{ts,js,svelte}"],
            exclude: [
                "src/**/*.test.{ts}",
                "src/**/*.d.ts",
                "src/lib/authService.ts",
            ],
            reportsDirectory: "./coverage",
            thresholds: {
                statements: 80,
                branches: 80,
                functions: 80,
                lines: 80,
            },
        },
    },
});
