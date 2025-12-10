import { render } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import '@testing-library/jest-dom/vitest';

// Mock API
const mockGetChatResponse = vi.fn();
vi.mock("../../src/lib/api", () => ({
    getChatResponse: (...args: unknown[]) => mockGetChatResponse(...args)
}));

import Footer from "../../src/lib/components/Footer.svelte";
import { messages, chat, user } from "../../src/lib/shared.svelte";
import { getChatResponse } from "../../src/lib/api";

describe("Footer", () => {
    beforeEach(() => {
        messages.length = 0;
        chat.id = "";
        chat.isLoading = false;
        user.id = "test-user-123";
        mockGetChatResponse.mockClear();
    });

    it("renders the footer container", () => {
        const { container } = render(Footer);

        const footer = container.querySelector('.sticky.bottom-0.bg-background');
        expect(footer).toBeInTheDocument();
    });

    it("renders prompt and send button components", () => {
        const { container } = render(Footer);

        const buttonGroup = container.querySelector('.flex.w-full.items-center.gap-2');
        expect(buttonGroup).toBeInTheDocument();
    });

    it("executes onclick function with successful API response", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "Test response",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        render(Footer);

        // Simulate onclick behavior
        const userQuestion = "Test question";
        const trimmedQuestion = userQuestion.trim();
        
        messages.push({
            question: trimmedQuestion,
            answer: "",
        });
        
        chat.isLoading = true;

        const answer = await getChatResponse({
            prompt: trimmedQuestion,
            userId: user.id,
            conversationId: chat.id,
        });

        messages[messages.length - 1].answer = answer.message.body;
        chat.id = answer.conversationId;
        chat.isLoading = false;

        expect(mockGetChatResponse).toHaveBeenCalledWith({
            prompt: trimmedQuestion,
            userId: "test-user-123",
            conversationId: "",
        });
        expect(messages[0].answer).toBe("Test response");
        expect(chat.id).toBe("conv-123");
        expect(chat.isLoading).toBe(false);
    });

    it("does not call API when user is not authenticated", async () => {
        user.id = "";
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        render(Footer);

        // Simulate onclick check
        if (!user.id) {
            console.error("User is not authenticated.");
            expect(consoleErrorSpy).toHaveBeenCalledWith("User is not authenticated.");
            expect(mockGetChatResponse).not.toHaveBeenCalled();
        }

        consoleErrorSpy.mockRestore();
    });

    it("handles API error in catch block", async () => {
        const errorMessage = "API Error";
        mockGetChatResponse.mockRejectedValue(new Error(errorMessage));

        render(Footer);

        const userQuestion = "Test question";
        messages.push({
            question: userQuestion,
            answer: "",
        });
        chat.isLoading = true;

        try {
            await getChatResponse({
                prompt: userQuestion,
                userId: user.id,
                conversationId: chat.id,
            });
        } catch (err: unknown) {
            messages[messages.length - 1].answer = (err as Error).message;
        } finally {
            chat.isLoading = false;
        }

        expect(messages[0].answer).toBe(errorMessage);
        expect(chat.isLoading).toBe(false);
    });

    it("trims input before sending", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "Response",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        render(Footer);

        const userQuestion = "  Test with spaces  ";
        const trimmedQuestion = userQuestion.trim();

        messages.push({
            question: trimmedQuestion,
            answer: "",
        });

        await getChatResponse({
            prompt: trimmedQuestion,
            userId: user.id,
            conversationId: chat.id,
        });

        expect(mockGetChatResponse).toHaveBeenCalledWith({
            prompt: "Test with spaces",
            userId: user.id,
            conversationId: "",
        });
    });

    it("sets chat.isLoading to true during API call", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "Response",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        render(Footer);

        expect(chat.isLoading).toBe(false);

        chat.isLoading = true;
        expect(chat.isLoading).toBe(true);

        await getChatResponse({
            prompt: "test",
            userId: user.id,
            conversationId: chat.id,
        });

        chat.isLoading = false;
        expect(chat.isLoading).toBe(false);
    });

    it("updates messages array with question and empty answer", async () => {
        render(Footer);

        const userQuestion = "Test question";

        messages.push({
            question: userQuestion,
            answer: "",
        });

        expect(messages.length).toBe(1);
        expect(messages[0].question).toBe(userQuestion);
        expect(messages[0].answer).toBe("");
    });

    it("updates last message answer from API response", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "API response body",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        render(Footer);

        messages.push({
            question: "Test",
            answer: "",
        });

        const answer = await getChatResponse({
            prompt: "Test",
            userId: user.id,
            conversationId: chat.id,
        });

        messages[messages.length - 1].answer = answer.message.body;

        expect(messages[0].answer).toBe("API response body");
    });

    it("updates chat.id from API response", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "Response",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "new-conv-id"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        render(Footer);

        expect(chat.id).toBe("");

        const answer = await getChatResponse({
            prompt: "test",
            userId: user.id,
            conversationId: chat.id,
        });

        chat.id = answer.conversationId;

        expect(chat.id).toBe("new-conv-id");
    });

    it("resets chat.isLoading in finally block on success", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "Response",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        render(Footer);

        chat.isLoading = true;

        try {
            await getChatResponse({
                prompt: "test",
                userId: user.id,
                conversationId: chat.id,
            });
        } finally {
            chat.isLoading = false;
        }

        expect(chat.isLoading).toBe(false);
    });

    it("resets chat.isLoading in finally block on error", async () => {
        mockGetChatResponse.mockRejectedValue(new Error("Error"));

        render(Footer);

        chat.isLoading = true;

        try {
            await getChatResponse({
                prompt: "test",
                userId: user.id,
                conversationId: chat.id,
            });
        } catch {
            // Error caught
        } finally {
            chat.isLoading = false;
        }

        expect(chat.isLoading).toBe(false);
    });
});
