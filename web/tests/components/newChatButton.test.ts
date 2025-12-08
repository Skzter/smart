import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, beforeEach } from "vitest";
import NewChatButton from "../../src/lib/components/NewChatButton.svelte";
import { chat, messages } from "$lib/shared.svelte";
import '@testing-library/jest-dom/vitest';

describe("NewChatButton", () => {
    beforeEach(() => {
        chat.id = "existing-chat-id";
        chat.isLoading = false;

        messages.length = 0;
        messages.push(
            { question: "Hello", answer: "Hi there!" }
        );
    });

    it("renders a button with 'New Chat' text", () => {
        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        expect(button).toBeInTheDocument();
    });

    it("displays Plus icon", () => {
        const { container } = render(NewChatButton);

        const svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();
    });

    it("displays 'New Chat' text", () => {
        render(NewChatButton);

        const text = screen.getByText("New Chat");
        expect(text).toBeInTheDocument();
    });

    it("resets chat.id to empty string when clicked", async () => {
        const user = userEvent.setup();
        chat.id = "some-chat-id";

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(chat.id).toBe("");
    });

    it("sets chat.isLoading to false when clicked", async () => {
        const user = userEvent.setup();
        chat.isLoading = true;

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(chat.isLoading).toBe(false);
    });

    it("clears messages array when clicked", async () => {
        const user = userEvent.setup();
        messages.push(
            { question: "Message 1", answer: "Response 1" },
            { question: "Message 2", answer: "Response 2" }
        );

        expect(messages.length).toBeGreaterThan(0);

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(messages.length).toBe(0);
    });

    it("resets all state when clicked", async () => {
        const user = userEvent.setup();

        chat.id = "old-chat-123";
        chat.isLoading = true;
        messages.push(
            { question: "Test message", answer: "Test response" }
        );

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(chat.id).toBe("");
        expect(chat.isLoading).toBe(false);
        expect(messages.length).toBe(0);
    });

    it("can be clicked multiple times", async () => {
        const user = userEvent.setup();

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });

        chat.id = "chat-1";
        messages.push({ question: "Message 1", answer: "Answer 1" });
        await user.click(button);

        expect(chat.id).toBe("");
        expect(messages.length).toBe(0);

        chat.id = "chat-2";
        messages.push({ question: "Message 2", answer: "Answer 2" });
        await user.click(button);

        expect(chat.id).toBe("");
        expect(messages.length).toBe(0);

        chat.id = "chat-3";
        messages.push({ question: "Message 3", answer: "Answer 3" });
        await user.click(button);

        expect(chat.id).toBe("");
        expect(messages.length).toBe(0);
    });

    it("works correctly when messages array is already empty", async () => {
        const user = userEvent.setup();
        messages.length = 0;

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(messages.length).toBe(0);
        expect(chat.id).toBe("");
        expect(chat.isLoading).toBe(false);
    });

    it("works correctly when chat.id is already empty", async () => {
        const user = userEvent.setup();
        chat.id = "";
        messages.push({ question: "Test", answer: "Test answer" });

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(chat.id).toBe("");
        expect(messages.length).toBe(0);
        expect(chat.isLoading).toBe(false);
    });

    it("is always enabled", () => {
        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        expect(button).not.toBeDisabled();
    });

    it("remains enabled even when chat is loading", () => {
        chat.isLoading = true;

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        expect(button).not.toBeDisabled();
    });

    it("preserves messages structure after clearing", async () => {
        const user = userEvent.setup();
        messages.push(
            { question: "Test 1", answer: "Response 1" }
        );

        render(NewChatButton);

        const button = screen.getByRole("button", { name: /new chat/i });
        await user.click(button);

        expect(messages.length).toBe(0);
        expect(Array.isArray(messages)).toBe(true);

        messages.push({ question: "New message", answer: "New answer" });
        expect(messages.length).toBe(1);
        expect(messages[0].question).toBe("New message");
    });
});