import eslint from "@eslint/js";
import tseslint from "typescript-eslint";
import svelte from "eslint-plugin-svelte";

export default [
    // Base JS/TS config
    {
        ...eslint.configs.recommended,
        ignores: [
            "dist/",
            "node_modules/",
            "public/",
            "src/assets/",
            ".vscode/*",
            "!.vscode/extensions.json",
            "*.config.*",
            "!eslint.config.js",
            "index.html",
            "package.json",
            "pnpm-*",
            "README.md",
            "tsconfig.*",
            "!tsconfig.json",
        ],
    },
    // TypeScript config
    {
        ...tseslint.configs.recommended,
        files: ["**/*.ts", "**/*.tsx"],
    },
    // Svelte config
    {
        files: ["**/*.svelte"],
        plugins: {
            svelte,
        },
        processor: svelte.processors[".svelte"],
        rules: {
            ...svelte.configs.recommended.rules,
        },
    },
];
