import { render, screen } from "@testing-library/svelte";
import { describe, it, expect, vi } from "vitest";
import Code from "../../src/components/Code.svelte";
import { tokenize } from "../../src/lib/syntaxHighlighting";

// Mock the tokenize function
vi.mock("../../src/lib/syntaxHighlighting", () => ({
    tokenize: vi.fn(),
}));

describe("Code Component", () => {
    beforeEach(() => {
        vi.mocked(tokenize).mockReset();
    });
    it("renders tokenized message correctly", () => {
        const mockTokens = [
            { type: "keyword", value: "const" },
            { type: "text", value: " " },
            { type: "variable", value: "foo" },
        ];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        render(Code, { props: { message: "const foo" } });

        const pre = screen.getByText(/const/).closest("pre");
        expect(pre).toBeInTheDocument();
        expect(pre?.textContent).toBe("const foo");
    });

    it("applies correct CSS classes to token spans", () => {
        const mockTokens = [
            { type: "keyword", value: "function" },
            { type: "operator", value: "=" },
        ];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        const { container } = render(Code, {
            props: { message: "function=" },
        });

        const keywordSpan = container.querySelector(".token-keyword");
        const operatorSpan = container.querySelector(".token-operator");

        expect(keywordSpan).toBeInTheDocument();
        expect(keywordSpan?.textContent).toBe("function");
        expect(operatorSpan).toBeInTheDocument();
        expect(operatorSpan?.textContent).toBe("=");
    });

    it("calls tokenize with the message prop", () => {
        const message = "Hello World";
        const mockTokens = [{ type: "text", value: message }];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        render(Code, { props: { message } });

        expect(tokenize).toHaveBeenCalledWith(message);
        expect(tokenize).toHaveBeenCalledTimes(1);
    });

    it("preserves whitespace in pre element", () => {
        const mockTokens = [{ type: "text", value: "line1\n  line2\t\ttab" }];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        const { container } = render(Code, {
            props: { message: "line1\n  line2\t\ttab" },
        });

        const pre = container.querySelector("pre");
        expect(pre).toHaveClass("whitespace-pre-wrap");
        expect(pre?.textContent).toBe("line1\n  line2\t\ttab");
    });

    it("handles empty message", () => {
        vi.mocked(tokenize).mockReturnValue([]);

        const { container } = render(Code, {
            props: { message: "" },
        });

        const pre = container.querySelector("pre");
        expect(pre).toBeInTheDocument();
        expect(pre?.textContent).toBe("");
    });

    it("applies all required Tailwind classes to pre", () => {
        const mockTokens = [{ type: "text", value: "test" }];
        vi.mocked(tokenize).mockReturnValue(mockTokens);

        const { container } = render(Code, {
            props: { message: "test" },
        });

        const pre = container.querySelector("pre");
        expect(pre).toHaveClass("font-sans");
        expect(pre).toHaveClass("text-base");
        expect(pre).toHaveClass("leading-relaxed");
        expect(pre).toHaveClass("whitespace-pre-wrap");
        expect(pre).toHaveClass("break-words");
    });

    it("renders multiple tokens with different types", () => {
        const mockTokens = [
            { type: "keyword", value: "import" },
            { type: "text", value: " " },
            { type: "string", value: '"module"' },
            { type: "text", value: " " },
            { type: "keyword", value: "from" },
            { type: "text", value: " " },
            { type: "string", value: '"path"' },
        ];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        const { container } = render(Code, {
            props: { message: 'import "module" from "path"' },
        });

        expect(container.querySelectorAll(".token-keyword")).toHaveLength(2);
        expect(container.querySelectorAll(".token-string")).toHaveLength(2);
        expect(container.querySelectorAll(".token-text")).toHaveLength(3);
    });

    it("reactively updates when message changes", () => {
        const mockTokens1 = [{ type: "text", value: "first" }];
        const mockTokens2 = [{ type: "text", value: "second" }];
        vi.mocked(tokenize)
            .mockReturnValueOnce(mockTokens1)
            .mockReturnValueOnce(mockTokens2);

        const { rerender } = render(Code, {
            props: { message: "first" },
        });

        expect(screen.getByText("first")).toBeInTheDocument();

        // Use rerender instead of component.$set
        rerender({ message: "second" });

        expect(screen.getByText("second")).toBeInTheDocument();
        expect(tokenize).toHaveBeenCalledTimes(2);
    });

    it("handles special HTML characters without escaping issues", () => {
        const mockTokens = [
            { type: "text", value: "<div>" },
            { type: "text", value: "&" },
            { type: "text", value: '"test"' },
        ];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        const { container } = render(Code, {
            props: { message: '<div>&"test"' },
        });

        const pre = container.querySelector("pre");
        expect(pre?.textContent).toBe('<div>&"test"');
    });
    it("applies correct CSS classes for token types", () => {
        const mockTokens = [
            { type: "keyword", value: "function" },
            { type: "functionCall", value: "myFunc" },
            { type: "string", value: '"hello"' },
            { type: "comment", value: "// comment" },
            { type: "identifier", value: "variable" },
            { type: "number", value: "42" },
            { type: "operator", value: "+" },
            { type: "punctuation", value: ";" },
            { type: "whitespace", value: " " },
        ];

        vi.mocked(tokenize).mockReturnValue(mockTokens);

        const { container } = render(Code, {
            props: { message: "function myFunc() { return 42; }" },
        });

        // Check each token has the correct class
        expect(container.querySelector(".token-keyword")).toBeInTheDocument();
        expect(
            container.querySelector(".token-functionCall"),
        ).toBeInTheDocument();
        expect(container.querySelector(".token-string")).toBeInTheDocument();
        expect(container.querySelector(".token-comment")).toBeInTheDocument();
        expect(
            container.querySelector(".token-identifier"),
        ).toBeInTheDocument();
        expect(container.querySelector(".token-number")).toBeInTheDocument();
        expect(container.querySelector(".token-operator")).toBeInTheDocument();
        expect(
            container.querySelector(".token-punctuation"),
        ).toBeInTheDocument();
        expect(
            container.querySelector(".token-whitespace"),
        ).toBeInTheDocument();

        // Verify content is correct
        expect(container.querySelector(".token-keyword")?.textContent).toBe(
            "function",
        );
        expect(container.querySelector(".token-number")?.textContent).toBe(
            "42",
        );
    });
    it("handles tokens with empty values", () => {
        vi.mocked(tokenize).mockReturnValue([{ type: "text", value: "" }]);

        const { container } = render(Code, {
            props: { message: "" },
        });

        const span = container.querySelector(".token-text");
        expect(span).toBeInTheDocument();
        expect(span?.textContent).toBe("");
    });

    it("handles tokens with special type", () => {
        vi.mocked(tokenize).mockReturnValue([{ type: "", value: "test" }]);

        const { container } = render(Code, {
            props: { message: "test" },
        });

        const span = container.querySelector('span[class="token-"]');
        expect(span).toBeInTheDocument();
    });
    it("handles token with type but no value content", () => {
        vi.mocked(tokenize).mockReturnValue([
            { type: "whitespace", value: "" },
        ]);

        const { container } = render(Code, {
            props: { message: "" },
        });

        const span = container.querySelector(".token-whitespace");
        expect(span).toBeInTheDocument();
    });
});
