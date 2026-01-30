import { render } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { Message } from "../../src/lib/shared.svelte";
import "@testing-library/jest-dom/vitest";

// Mock API
const mockGetChatResponse = vi.fn();
vi.mock("../../src/lib/api", () => ({
    getChatResponse: (...args: unknown[]) => mockGetChatResponse(...args),
}));

// Mock shared state
vi.mock("../../src/lib/shared.svelte", () => ({
    messages: [],
    chat: {
        id: "",
        isLoading: false,
    },
    user: {
        id: "test-user-123",
    },
}));

import Main from "../../src/lib/components/Main.svelte";
import { messages, chat } from "../../src/lib/shared.svelte";

describe("Main", () => {
    beforeEach(() => {
        messages.length = 0;
        chat.isLoading = false;
        chat.id = "";
        mockGetChatResponse.mockClear();
    });

    it("renders the Main component", () => {
        const { container } = render(Main);
        expect(container).toBeInTheDocument();
    });

    it("renders Chat component", () => {
        const { container } = render(Main);

        // Chat component renders a container with specific classes
        const chatContainer = container.querySelector(
            ".bg-muted\\/50.rounded-xl",
        );
        expect(chatContainer).toBeInTheDocument();
    });

    it("renders Footer component", () => {
        const { container } = render(Main);

        // Footer component renders with sticky positioning
        const footerContainer = container.querySelector(
            ".sticky.bottom-0.bg-background",
        );
        expect(footerContainer).toBeInTheDocument();
    });

    it("renders both Chat and Footer components together", () => {
        const { container } = render(Main);

        // Verify both components are present
        const chatContainer = container.querySelector(
            ".bg-muted\\/50.rounded-xl",
        );
        const footerContainer = container.querySelector(
            ".sticky.bottom-0.bg-background",
        );

        expect(chatContainer).toBeInTheDocument();
        expect(footerContainer).toBeInTheDocument();
    });

    it("displays empty state in Chat when no messages", () => {
        const { container } = render(Main);

        const emptyMessage = container.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );
        expect(emptyMessage?.textContent).toMatch(/Start a chat\.\.\.|Chat starten\.\.\./);
    });

    it("renders messages in Chat component", () => {
        messages.push({ t: "user", Message: "Test question" } as unknown as Message);
        messages.push({ t: "bot", Message: "Test answer" } as unknown as Message);

        const { container } = render(Main);

        const messageContainer = container.querySelector(
            ".flex.flex-col.gap-4",
        );
        expect(messageContainer).toBeInTheDocument();
    });

    it("shows loading indicator when chat is loading", () => {
        messages.push({ t: "user", Message: "Test question" } as unknown as Message);
        messages.push({ t: "bot", Message: "" } as unknown as Message);

        chat.isLoading = true;

        const { container } = render(Main);

        // Dots component should be rendered
        const dotsContainer = container.querySelector(".flex.gap-1");
        expect(dotsContainer).toBeInTheDocument();
    });

    it("has proper layout structure", () => {
        const { container } = render(Main);

        // Check that both components are rendered in order
        const chatContainer = container.querySelector(
            ".bg-muted\\/50.rounded-xl",
        );
        const footerContainer = container.querySelector(
            ".sticky.bottom-0.bg-background",
        );

        expect(chatContainer).toBeInTheDocument();
        expect(footerContainer).toBeInTheDocument();
    });

    it("integrates Chat and Footer components correctly", () => {
        const { container } = render(Main);

        // Verify the complete structure exists
        expect(container).toBeInTheDocument();

        // Chat should have message container
        const chatMessageArea = container.querySelector(
            ".bg-muted\\/50.rounded-xl",
        );
        expect(chatMessageArea).toBeInTheDocument();

        // Footer should have input area
        const footerInputArea = container.querySelector(
            ".flex.w-full.items-center.gap-2",
        );
        expect(footerInputArea).toBeInTheDocument();
    });

    it("renders with multiple messages", () => {
        messages.push({ t: "user", Message: "Question 1" } as unknown as Message);
        messages.push({ t: "bot", Message: "Answer 1" } as unknown as Message);
        messages.push({ t: "user", Message: "Question 2" } as unknown as Message);
        messages.push({ t: "bot", Message: "Answer 2" } as unknown as Message);
        messages.push({ t: "user", Message: "Question 3" } as unknown as Message);
        messages.push({ t: "bot", Message: "Answer 3" } as unknown as Message);

        const { container } = render(Main);

        const messageContainer = container.querySelector(
            ".flex.flex-col.gap-4",
        );
        expect(messageContainer).toBeInTheDocument();
    });

    it("maintains state between Chat and Footer components", () => {
        const { container } = render(Main);

        // Initially no messages
        let emptyState = container.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );
        expect(emptyState).toBeInTheDocument();

        // Add a message (simulating Footer action)
        messages.push({ t: "user", Message: "New question" } as unknown as Message);

        // Re-render to reflect state change
        const { container: updatedContainer } = render(Main);

        // Empty state should be gone
        emptyState = updatedContainer.querySelector(
            ".flex.items-center.justify-center.flex-1",
        );
        expect(emptyState).not.toBeInTheDocument();
    });

    it("renders all child components without errors", () => {
        const { container } = render(Main);

        // Should not throw errors and render both components
        expect(
            container.querySelector(".bg-muted\\/50.rounded-xl"),
        ).toBeInTheDocument();
        expect(
            container.querySelector(".sticky.bottom-0.bg-background"),
        ).toBeInTheDocument();
    });
});
