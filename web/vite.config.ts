import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import path from "path";
import mkcert from "vite-plugin-mkcert";

// https://vite.dev/config/
export default defineConfig({
    server: {
        https: true,
    },
    plugins: [tailwindcss(), svelte(), mkcert()],
    resolve: {
        alias: {
            $lib: path.resolve("./src/lib"),
            $types: path.resolve("./src/types"),
        },
    },
});
