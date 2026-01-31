import { render } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock shared state BEFORE importing BotMessage
vi.mock("../../src/lib/shared.svelte", () => ({
    user: { id: "user123" },
    chat: { id: "chat456" },
}));

// Mock Monaco Editor BEFORE importing components
vi.mock("monaco-editor", () => ({
    editor: {
        create: vi.fn(() => ({
            getValue: vi.fn(() => ""),
            setValue: vi.fn(),
            getModel: vi.fn(() => ({
                getValue: vi.fn(() => ""),
                setValue: vi.fn(),
                onDidChangeContent: vi.fn(() => ({ dispose: vi.fn() })),
            })),
            layout: vi.fn(),
            dispose: vi.fn(),
            onDidChangeModelContent: vi.fn(() => ({ dispose: vi.fn() })),
        })),
        createModel: vi.fn(),
        setModelLanguage: vi.fn(),
    },
}));

// Mock clipboard BEFORE importing the component
const mockWriteText = vi.fn().mockResolvedValue(undefined);
Object.defineProperty(navigator, "clipboard", {
    value: {
        writeText: mockWriteText,
    },
    writable: true,
    configurable: true,
});

import BotMessage from "../../src/lib/components/BotMessage.svelte";

describe.skip("BotMessage TODO: fix this test", () => {
    beforeEach(() => {
        mockWriteText.mockClear();
        mockWriteText.mockResolvedValue(undefined);
    });

    it("renders bot icon and message container", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: { t: "error", Message: "Test message" },
            },
        });

        const botIcon = container.querySelector("svg");
        const messageContainer = container.querySelector(
            ".bg-muted.rounded-2xl",
        );

        expect(botIcon).toBeInTheDocument();
        expect(messageContainer).toBeInTheDocument();
    });

    it("renders regular text message correctly", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: { t: "error", Message: "This is a regular message" },
            },
        });

        const textDiv = container.querySelector(".px-4.py-2");
        expect(textDiv).toBeInTheDocument();
        expect(textDiv?.textContent).toBe("This is a regular message");
    });

    it("detects and renders Playwright code as Code component", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: {
                    t: "generation",
                    Message: "import { test } from '@playwright/test';",
                },
            },
        });

        // When message contains @playwright, MonacoEditor should be rendered
        // Check that regular text div is NOT present
        const textDiv = container.querySelector(".px-4.py-2.wrap-break-word");
        expect(textDiv).not.toBeInTheDocument();
    });

    it("does not render Code component for regular text", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: { t: "validation", Message: "Just a regular message" },
            },
        });

        // Regular text should be rendered in text div
        const textDiv = container.querySelector(".px-4.py-2.wrap-break-word");
        expect(textDiv).toBeInTheDocument();
    });

    it("renders TestButtons component", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: { t: "error", Message: "Test" },
            },
        });

        const testButtons = container.querySelector(".border-b");
        expect(testButtons).toBeInTheDocument();
    });

    it("handles empty message in text mode", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: { t: "error", Message: "" },
            },
        });

        const textDiv = container.querySelector(".px-4.py-2");
        expect(textDiv).toBeInTheDocument();
        expect(textDiv?.textContent).toBe("");
    });

    it("preserves whitespace in text messages", () => {
        const { container } = render(BotMessage, {
            props: {
                msg: { t: "error", Message: "Line 1\n  Line 2\n    Line 3" },
            },
        });

        const textDiv = container.querySelector(".whitespace-pre-wrap");
        expect(textDiv).toBeInTheDocument();
        expect(textDiv?.textContent).toContain("\n");
    });
});
