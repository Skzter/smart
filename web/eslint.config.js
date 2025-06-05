// empfohlene config con typescript-eslint.io
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
    eslint.configs.recommended,
    tseslint.configs.recommended,
    {
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
);
