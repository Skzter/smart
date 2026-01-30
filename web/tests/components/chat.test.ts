import { render } from "@testing-library/svelte";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import "@testing-library/jest-dom/vitest";

vi.mock("../../src/lib/shared.svelte", () => ({
    messages: [],
    chat: {
        id: "",
        isLoading: false,
        groups: [],
    },
    GroupsState: {
        items: [],
        isLoading: false,
        error: "",
    },
}));

import Chat from "../../src/lib/components/Chat.svelte";
import { messages, chat } from "../../src/lib/shared.svelte";

describe("Chat", () => {
    beforeEach(() => {
        messages.length = 0;
        chat.isLoading = false;
        chat.groups = [];
        chat.id = "";

        vi.spyOn(window, "confirm").mockReturnValue(true);
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it("displays empty state when no messages exist", () => {
        const { container } = render(Chat);

        const emptyMessage = container.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );

        expect(emptyMessage).toBeInTheDocument();
        expect(emptyMessage?.textContent).toBe("Chat starten...");
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
            t: "user",
            Message: "Test message",
        });

        const { container } = render(Chat);

        const messageContainer = container.querySelector(
            ".flex.flex-col.gap-4",
        );

        expect(messageContainer).toBeInTheDocument();
    });

    it("shows loading indicator when chat is loading", () => {
        messages.push({
            t: "user",
            Message: "Test message",
        });

        chat.isLoading = true;

        const { container } = render(Chat);

        // Dots component → simple structural assertion
        const dotsContainer = container.querySelector(".flex.gap-1");
        expect(dotsContainer).toBeInTheDocument();
    });

    it("does not show empty state when messages exist", () => {
        messages.push({
            t: "user",
            Message: "Test message",
        });

        const { container } = render(Chat);

        const emptyMessage = container.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );

        expect(emptyMessage).not.toBeInTheDocument();
    });
});
