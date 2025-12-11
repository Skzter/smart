import { render } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import '@testing-library/jest-dom/vitest';

import Code from "../../src/lib/components/Code.svelte";

describe("Code", () => {
    it("renders a pre element", () => {
        const { container } = render(Code, {
            props: {
                code: "const test = 'hello';",
            },
        });

        const preElement = container.querySelector('pre');
        expect(preElement).toBeInTheDocument();
    });

    it("renders code with syntax highlighting", () => {
        const { container } = render(Code, {
            props: {
                code: "const test = 'hello';",
            },
        });

        const tokens = container.querySelectorAll('span[class^="token-"]');
        expect(tokens.length).toBeGreaterThan(0);
    });

    it("preserves whitespace and line breaks", () => {
        const { container } = render(Code, {
            props: {
                code: "line 1\n  line 2",
            },
        });

        const preElement = container.querySelector('pre.whitespace-pre-wrap');
        expect(preElement).toBeInTheDocument();
    });

    it("renders empty code", () => {
        const { container } = render(Code, {
            props: {
                code: "",
            },
        });

        const preElement = container.querySelector('pre');
        expect(preElement).toBeInTheDocument();
    });

    it("tokenizes keywords correctly", () => {
        const { container } = render(Code, {
            props: {
                code: "const import function",
            },
        });

        const keywordTokens = container.querySelectorAll('span.token-keyword');
        expect(keywordTokens.length).toBeGreaterThan(0);
    });

    it("renders multiple tokens in each block", () => {
        const { container } = render(Code, {
            props: {
                code: "const x = 123;",
            },
        });

        const allTokens = container.querySelectorAll('span[class^="token-"]');
        expect(allTokens.length).toBeGreaterThan(1);
        
        // Verify each token has content
        allTokens.forEach(token => {
            expect(token.textContent).toBeTruthy();
        });
    });

    it("handles code with various token types", () => {
        const { container } = render(Code, {
            props: {
                code: 'const name = "test"; // comment',
            },
        });

        const spans = container.querySelectorAll('span');
        expect(spans.length).toBeGreaterThan(0);
    });

    it("tokenizes different code correctly", () => {
        const { container } = render(Code, {
            props: {
                code: "function test() { return 42; }",
            },
        });

        const tokens = container.querySelectorAll('span[class^="token-"]');
        expect(tokens.length).toBeGreaterThan(5);
    });

    it("handles complex code with multiple statement types", () => {
        const complexCode = `async function getData() {
    const response = await fetch('/api');
    return response.json();
}`;
        const { container } = render(Code, {
            props: {
                code: complexCode,
            },
        });

        const tokens = container.querySelectorAll('span[class^="token-"]');
        expect(tokens.length).toBeGreaterThan(10);
        
        // Should have keywords
        const keywords = container.querySelectorAll('span.token-keyword');
        expect(keywords.length).toBeGreaterThan(0);
    });

    it("renders token values correctly", () => {
        const { container } = render(Code, {
            props: {
                code: "const x = 123;",
            },
        });

        const tokens = container.querySelectorAll('span[class^="token-"]');
        
        // Verify each token has both type class and value content
        tokens.forEach(token => {
            expect(token.className).toMatch(/^token-/);
            expect(token.textContent).not.toBe("");
        });
    });

    it("renders all token types and values in each iteration", () => {
        const { container } = render(Code, {
            props: {
                code: 'let name = "value"; // comment\nconst num = 42;',
            },
        });

        const allSpans = container.querySelectorAll('span[class^="token-"]');
        expect(allSpans.length).toBeGreaterThan(0);
        
        // Each span should have token type as class and value as content
        allSpans.forEach(span => {
            const className = span.className;
            const textContent = span.textContent;
            
            expect(className).toContain('token-');
            expect(textContent).toBeDefined();
        });
    });
});
