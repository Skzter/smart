import { render, screen } from "@testing-library/svelte";
import { describe, it, expect } from "vitest";
import UserMessage from "../../src/lib/components/UserMessage.svelte";
import "@testing-library/jest-dom/vitest";

describe("UserMessage", () => {
    it("renders the message text", () => {
        const message = "Hello, this is a test message";
        render(UserMessage, { props: { message } });

        expect(screen.getByText(message)).toBeInTheDocument();
    });

    it("displays User icon", () => {
        const { container } = render(UserMessage, {
            props: { message: "Test" },
        });

        const svg = container.querySelector("svg");
        expect(svg).toBeInTheDocument();
        expect(svg).toHaveClass("h-7", "w-7", "text-primary-foreground");
    });

    it("applies correct message bubble styling", () => {
        const { container } = render(UserMessage, {
            props: { message: "Test message" },
        });

        const messageBubble = container.querySelector(".bg-primary");
        expect(messageBubble).toHaveClass(
            "text-primary-foreground",
            "rounded-2xl",
            "rounded-tr-sm",
            "px-4",
            "py-2",
            "max-w-[80%]",
        );
    });

    it("applies correct avatar styling", () => {
        const { container } = render(UserMessage, {
            props: { message: "Test" },
        });

        const avatar = container.querySelector(".h-8.w-8");
        expect(avatar).toHaveClass(
            "shrink-0",
            "rounded-full",
            "bg-primary",
            "flex",
            "items-center",
            "justify-center",
        );
    });

    it("aligns message to the right", () => {
        const { container } = render(UserMessage, {
            props: { message: "Test" },
        });

        const wrapper = container.querySelector(".flex");
        expect(wrapper).toHaveClass("justify-end", "gap-2", "items-start");
    });

    it("handles multiline messages with whitespace", () => {
        const message = "Line 1\nLine 2\nLine 3";
        const { container } = render(UserMessage, { props: { message } });

        const messageBubble = container.querySelector(".whitespace-pre-wrap");
        expect(messageBubble).toBeInTheDocument();
        expect(messageBubble?.textContent).toBe(message);
    });

    it("handles empty message", () => {
        const { container } = render(UserMessage, {
            props: { message: "" },
        });

        const messageBubble = container.querySelector(".bg-primary");
        expect(messageBubble).toBeInTheDocument();
        expect(messageBubble?.textContent).toBe("");
    });
});
