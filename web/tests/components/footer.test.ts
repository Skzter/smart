import { render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import '@testing-library/jest-dom/vitest';

// Mock API
const mockGetChatResponse = vi.fn();
const mockGetTemplate = vi.fn();
vi.mock("../../src/lib/api", () => ({
    getChatResponse: (...args: unknown[]) => mockGetChatResponse(...args),
    getTemplate: (...args: unknown[]) => mockGetTemplate(...args)
}));

// Mock toast
vi.mock("svelte-sonner", () => ({
    toast: {
        error: vi.fn(),
    },
}));

import Footer from "../../src/lib/components/Footer.svelte";
import { messages, chat, user } from "../../src/lib/shared.svelte";

describe("Footer", () => {
    beforeEach(() => {
        messages.length = 0;
        chat.id = "";
        chat.isLoading = false;
        user.id = "test-user-123";
        mockGetChatResponse.mockClear();
        mockGetTemplate.mockClear();
        mockGetTemplate.mockResolvedValue("Default template");
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

    it("renders textarea for input", async () => {
        render(Footer);
        
        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });
    });

    it("sends message with successful API response", async () => {
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

        const userSetup = userEvent.setup();
        render(Footer);
        
        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test question");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(mockGetChatResponse).toHaveBeenCalledWith({
                prompt: "Test question",
                userId: "test-user-123",
                conversationId: "",
            });
            expect(messages[0].answer).toBe("Test response");
            expect(chat.id).toBe("conv-123");
        });
    });

    it("does not call API when user is not authenticated", async () => {
        user.id = "";
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test question");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(consoleErrorSpy).toHaveBeenCalledWith("User is not authenticated.");
        });
        expect(mockGetChatResponse).not.toHaveBeenCalled();
        consoleErrorSpy.mockRestore();
    });

    it("handles API error and shows error message", async () => {
        const errorMessage = "API Error";
        mockGetChatResponse.mockRejectedValue(new Error(errorMessage));

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test question");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(messages[0].answer).toBe(errorMessage);
            expect(chat.isLoading).toBe(false);
        });
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

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "  Test with spaces  ");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(mockGetChatResponse).toHaveBeenCalledWith({
                prompt: "Test with spaces",
                userId: user.id,
                conversationId: "",
            });
        });
    });

    it("sets chat.isLoading to true during API call and resets after", async () => {
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
        
        let resolveFn: (value: unknown) => void;
        const delayedPromise = new Promise((resolve) => {
            resolveFn = resolve;
        });
        
        mockGetChatResponse.mockReturnValue(delayedPromise);

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        expect(chat.isLoading).toBe(false);

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "test");
        
        const button = screen.getByRole("button");
        const clickPromise = userSetup.click(button);
        
        // Should be loading now
        await waitFor(() => {
            expect(chat.isLoading).toBe(true);
        }, { timeout: 3000 });

        // Resolve the API call
        resolveFn!(mockResponse);
        await clickPromise;

        // Should not be loading anymore
        await waitFor(() => {
            expect(chat.isLoading).toBe(false);
        });
    });

    it("clears input after sending message", async () => {
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

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...") as HTMLTextAreaElement;
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test question");
        
        expect(textarea.value).toBe("Test question");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            // The onclick function clears input by setting it to ""
            expect(textarea.value).toBe("");
        });
    });

    it("updates messages array with question", async () => {
        const mockResponse = {
            message: {
                id: "msg-1",
                body: "Answer",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test question");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(messages.length).toBe(1);
            expect(messages[0].question).toBe("Test question");
        });
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

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(messages[0].answer).toBe("API response body");
        });
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

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        expect(chat.id).toBe("");

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "test");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(chat.id).toBe("new-conv-id");
        });
    });

    it("handles existing conversationId in subsequent calls", async () => {
        chat.id = "existing-conv-123";
        
        const mockResponse = {
            message: {
                id: "msg-2",
                body: "Follow-up response",
                role: "assistant",
                createdAt: Date.now()
            },
            userId: "test-user-123",
            conversationId: "existing-conv-123"
        };
        mockGetChatResponse.mockResolvedValue(mockResponse);

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Follow-up question");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(mockGetChatResponse).toHaveBeenCalledWith({
                prompt: "Follow-up question",
                userId: "test-user-123",
                conversationId: "existing-conv-123",
            });
        });
    });

    it("creates correct API request object", async () => {
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

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test");
        
        const button = screen.getByRole("button");
        await userSetup.click(button);

        await waitFor(() => {
            expect(mockGetChatResponse).toHaveBeenCalledWith(
                expect.objectContaining({
                    prompt: "Test",
                    userId: "test-user-123",
                    conversationId: "",
                })
            );
        });
    });

    it("supports Enter key to send message", async () => {
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

        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test with enter{Enter}");

        await waitFor(() => {
            expect(mockGetChatResponse).toHaveBeenCalledWith(
                expect.objectContaining({
                    prompt: "Test with enter",
                    userId: "test-user-123",
                })
            );
        });
    });

    it("does not send message when Shift+Enter is pressed", async () => {
        const userSetup = userEvent.setup();
        render(Footer);

        await waitFor(() => {
            const textarea = screen.getByPlaceholderText("Send a message...");
            expect(textarea).toBeInTheDocument();
        });

        const textarea = screen.getByPlaceholderText("Send a message...");
        await userSetup.clear(textarea);
        await userSetup.type(textarea, "Test");
        await userSetup.keyboard("{Shift>}{Enter}{/Shift}");

        // Should not call API
        expect(mockGetChatResponse).not.toHaveBeenCalled();
        
        // Text should still be in textarea (for multiline)
        expect((textarea as HTMLTextAreaElement).value).toContain("Test");
    });
});
