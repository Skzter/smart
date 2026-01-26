import { render } from "@testing-library/svelte";
import { describe, it, expect, beforeEach, vi } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock shared state
vi.mock("../../src/lib/shared.svelte", () => ({
    messages: [],
    chat: {
        id: "",
        isLoading: false,
    },
}));

import Chat from "../../src/lib/components/Chat.svelte";
import { messages, chat } from "../../src/lib/shared.svelte";

describe("Chat", () => {
    beforeEach(() => {
        // Reset state before each test
        messages.length = 0;
        chat.isLoading = false;
    });

    it("displays empty state when no messages", () => {
        const { container } = render(Chat);

        const emptyMessage = container.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );
        expect(emptyMessage?.textContent).toBe("Start a chat...");
    });

    it("renders chat container with correct styling", () => {
        const { container } = render(Chat);

        const chatContainer = container.querySelector(
            ".bg-muted\\/50.rounded-xl",
        );
        expect(chatContainer).toBeInTheDocument();
    });

    it("renders messages when they exist", () => {
        messages.push({
            question: "Test question",
            answer: "Test answer",
        });

        const { container } = render(Chat);

        const messageContainer = container.querySelector(
            ".flex.flex-col.gap-4",
        );
        expect(messageContainer).toBeInTheDocument();
    });

    it("shows loading indicator when chat is loading", () => {
        messages.push({
            question: "Test question",
            answer: "",
        });
        chat.isLoading = true;

        const { container } = render(Chat);

        // Dots component should be rendered (looking for the animated dots)
        const dotsContainer = container.querySelector(".flex.gap-1");
        expect(dotsContainer).toBeInTheDocument();
    });

    it("does not show empty state when messages exist", () => {
        messages.push({
            question: "Test",
            answer: "Answer",
        });

        const { container } = render(Chat);

        const emptyMessage = container.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );
        expect(emptyMessage).not.toBeInTheDocument();
    });
});
