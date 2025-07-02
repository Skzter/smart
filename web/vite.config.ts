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
        coverage: {
            provider: "v8",
            reporter: ["text", "json", "html"],
            include: ["src/{lib,components}/**/*.{ts,js,svelte}"],
            exclude: ["src/**/*.test.{ts}", "src/**/*.d.ts"],
        },
    },
});
