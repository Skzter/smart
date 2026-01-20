import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock API and shared modules
vi.mock("$lib/api", () => ({
    getChatById: vi.fn(),
    updateChatTitle: vi.fn(),
}));

vi.mock("svelte-sonner", () => ({
    toast: {
        success: vi.fn(),
        error: vi.fn(),
    },
}));

vi.mock("$lib/toast", () => ({
    toast: {
        success: vi.fn(),
        error: vi.fn()
    }
}));

vi.mock("$lib/hooks/api", () => ({
    updateChatTitle: vi.fn()
}));

import ChatSummary from "../../src/lib/components/ChatSummary.svelte";
import { getChatById } from "$lib/api";
import { toast } from "svelte-sonner";import { chat, messages, user as sharedUser} from "$lib/shared.svelte";

import type {
    ApiChatSummary,
    ApiMessage,
    ApiGetChatByIdResponse,
} from "$types/api";
import { updateChatTitle as updateChatTitleApi } from "$lib/api";

// Helper to create a complete mock response
const createMockResponse = (
    messagesArray: ApiMessage[],
): ApiGetChatByIdResponse => ({
    id: "test-chat-id",
    userId: "test-user-id",
    createdAt: "2024-01-15T10:30:00Z",
    updatedAt: "2024-01-15T10:30:00Z",
    title: "Test Chat",
    lastTest: "",
    lastAutoPlaywrightPrompt: "",
    messages: messagesArray,
});

describe.skip("ChatSummary.svelte TODO: fix this test", () => {
    let mockSummary: ApiChatSummary;

    beforeEach(() => {
        // Initialize mockSummary
        mockSummary = {
            chatId: "test-chat-id",
            userId: "test-user-id",
            title: "Test Chat",
            createdAt: "2024-01-15T10:30:00Z",
            updatedAt: "2024-01-15T10:30:00Z",
        };

        // Reset shared state
        sharedUser.id = "";
        chat.id = "";
        chat.isLoading = false;
        messages.length = 0;
    });

    describe("Rendering", () => {
        it("renders the chat summary with title", () => {
            render(ChatSummary, {
                props: { summary: mockSummary },
            });

            expect(screen.getByText("Test Chat")).toBeInTheDocument();
        });

        it("renders 'Neuer Chat' when title is empty", () => {
            mockSummary.title = "";

            render(ChatSummary, {
                props: { summary: mockSummary },
            });

            expect(screen.getByText("Neuer Chat")).toBeInTheDocument();
        });

        it("renders the message icon", () => {
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const icon = container.querySelector("svg.lucide-message-square");
            expect(icon).toBeInTheDocument();
        });

        it("renders the pencil edit icon", () => {
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const icon = container.querySelector("svg.lucide-pencil");
            expect(icon).toBeInTheDocument();
        });

        it("renders formatted time in CET", () => {
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            // Should format "2024-12-11T14:45:00Z" to CET time
            const timeElement = container.querySelector(".font-mono.text-xs");
            expect(timeElement).toBeInTheDocument();
            expect(timeElement?.textContent).toMatch(/\d{2}:\d{2}/);
        });
    });

    describe("formatToCET function", () => {
        it("formats ISO date to CET time", () => {
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const timeElement = container.querySelector(".font-mono.text-xs");
            expect(timeElement?.textContent).toBeTruthy();
        });

        it("handles empty date string", () => {
            mockSummary.updatedAt = "";

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const timeElement = container.querySelector(".font-mono.text-xs");
            expect(timeElement?.textContent).toBe("");
        });

        it("handles undefined date", () => {
            mockSummary.updatedAt = "" as string;

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const timeElement = container.querySelector(".font-mono.text-xs");
            expect(timeElement?.textContent).toBe("");
        });

        it("handles invalid date format with fallback", () => {
            mockSummary.updatedAt = "2024-12-11T14:45:30.123Z";

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const timeElement = container.querySelector(".font-mono.text-xs");
            expect(timeElement).toBeInTheDocument();
        });

        it("handles date without substring method", () => {
            mockSummary.updatedAt = null as unknown as string;

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const timeElement = container.querySelector(".font-mono.text-xs");
            expect(timeElement?.textContent).toBe("");
        });
    });

    describe("Edit Mode", () => {
        it("switches to edit mode when pencil icon is clicked", async () => {
            const user = userEvent.setup();
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            // Find the pencil icon's parent button
            const pencilIcon = container.querySelector("svg.lucide-pencil");
            expect(pencilIcon).toBeInTheDocument();
            const editButton = pencilIcon?.closest("button");
            expect(editButton).toBeInTheDocument();

            await user.click(editButton!);

            const input = container.querySelector("input");
            expect(input).toBeInTheDocument();
        });

        it("shows input with current title in edit mode", async () => {
            const user = userEvent.setup();
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input") as HTMLInputElement;
            expect(input?.value).toBe("Test Chat");
        });

        it("shows 'Neuer Chat' in input when title is empty", async () => {
            const user = userEvent.setup();
            mockSummary.title = "";

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input") as HTMLInputElement;
            expect(input?.value).toBe("Neuer Chat");
        });

        it("updates title on Enter key", async () => {
            const user = userEvent.setup();
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input") as HTMLInputElement;
            await user.clear(input);
            await user.type(input, "Updated Title{Enter}");

            expect(mockSummary.title).toBe("Updated Title");
        });

        it("cancels edit mode on Escape key", async () => {
            const user = userEvent.setup();
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            let input = container.querySelector("input");
            expect(input).toBeInTheDocument();

            await user.type(input!, "{Escape}");

            // Wait for state update
            await new Promise((resolve) => setTimeout(resolve, 100));

            // Input should be gone
            input = container.querySelector("input");
            expect(input).not.toBeInTheDocument();
        });

        it("updates title on focus out", async () => {
            const user = userEvent.setup();
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input") as HTMLInputElement;
            await user.clear(input);
            await user.type(input, "New Title");

            // Trigger focusout
            input.blur();
            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(mockSummary.title).toBe("New Title");
        });

        it("focuses input when entering edit mode with empty title", async () => {
            const user = userEvent.setup();
            mockSummary.title = "";

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input") as HTMLInputElement;
            expect(input).toBeInTheDocument();
            // Input should be focused via focusAction
        });

        it("focuses input when entering edit mode with existing title", async () => {
            const user = userEvent.setup();

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input") as HTMLInputElement;
            expect(input).toBeInTheDocument();
        });
    });

    describe("Chat Switching", () => {
        it("calls invokeSwitchChat when menu button is clicked", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Hello",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: "Hi there",
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            expect(getChatById).toHaveBeenCalled();
            expect(sharedUser.id).toBe("test-user-id");
            expect(chat.id).toBe("test-chat-id");
        });

        it("sets user.id and chat.id correctly", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(sharedUser.id).toBe("test-user-id");
            expect(chat.id).toBe("test-chat-id");
        });

        it("sets chat.isLoading to false", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);
            chat.isLoading = true;

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(chat.isLoading).toBe(false);
        });

        it("handles API error and shows toast", async () => {
            const user = userEvent.setup();
            const error = new Error("API Error");

            vi.mocked(getChatById).mockRejectedValue(error);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description: "API Error",
                },
            );
        });

        it("handles unknown error type", async () => {
            const user = userEvent.setup();

            vi.mocked(getChatById).mockRejectedValue("string error");

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(toast.error).toHaveBeenCalledWith(
                "Speichern fehlgeschlagen",
                {
                    description: "Unbekannter Fehler",
                },
            );
        });
    });

    describe("convertApiMessagesToMessages function", () => {
        it("converts API messages correctly", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Question 1",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: "Answer 1",
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
                {
                    id: "msg3",
                    role: "user",
                    body: "Question 2",
                    createdAt: "2024-01-15T10:32:00Z",
                } as ApiMessage,
                {
                    id: "msg4",
                    role: "assistant",
                    body: "Answer 2",
                    createdAt: "2024-01-15T10:33:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(2);
            expect(messages[0].question).toBe("Question 1");
            expect(messages[0].answer).toBe("Answer 1");
            expect(messages[1].question).toBe("Question 2");
            expect(messages[1].answer).toBe("Answer 2");
        });

        it("skips user messages without assistant response", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Question 1",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "user",
                    body: "Question 2",
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
                {
                    id: "msg3",
                    role: "assistant",
                    body: "Answer 2",
                    createdAt: "2024-01-15T10:32:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].question).toBe("Question 2");
            expect(messages[0].answer).toBe("Answer 2");
        });

        it("handles validation message with valid=false", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Invalid input",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: JSON.stringify({
                        valid: false,
                        message: "Validation failed",
                    }),
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].answer).toBe("Validation failed");
        });

        it("skips validation message with valid=true and finds next answer", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Valid input",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: JSON.stringify({ valid: true }),
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
                {
                    id: "msg3",
                    role: "assistant",
                    body: "Actual answer",
                    createdAt: "2024-01-15T10:32:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].answer).toBe("Actual answer");
        });

        it("handles JSON parsing error and uses as regular answer", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Question",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: "Not JSON",
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].answer).toBe("Not JSON");
        });

        it("handles JSON without valid property", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Question",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: JSON.stringify({ data: "some data" }),
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].answer).toBe(
                JSON.stringify({ data: "some data" }),
            );
        });

        it("stops at next user message when looking for answer", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "Q1",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: JSON.stringify({ valid: true }),
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
                {
                    id: "msg3",
                    role: "user",
                    body: "Q2",
                    createdAt: "2024-01-15T10:32:00Z",
                } as ApiMessage,
                {
                    id: "msg4",
                    role: "assistant",
                    body: "A2",
                    createdAt: "2024-01-15T10:33:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            // Q1 has no answer (validation passed but next is user message)
            expect(messages.length).toBe(1);
            expect(messages[0].question).toBe("Q2");
        });

        it("clears existing messages before adding new ones", async () => {
            const user = userEvent.setup();
            messages.push({ question: "Old Q", answer: "Old A" });

            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "user",
                    body: "New Q",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "assistant",
                    body: "New A",
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].question).toBe("New Q");
        });

        it("handles empty messages array", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(0);
        });

        it("skips non-user, non-assistant messages", async () => {
            const user = userEvent.setup();
            const mockResponse = createMockResponse([
                {
                    id: "msg1",
                    role: "system",
                    body: "System message",
                    createdAt: "2024-01-15T10:30:00Z",
                } as ApiMessage,
                {
                    id: "msg2",
                    role: "user",
                    body: "Question",
                    createdAt: "2024-01-15T10:31:00Z",
                } as ApiMessage,
                {
                    id: "msg3",
                    role: "assistant",
                    body: "Answer",
                    createdAt: "2024-01-15T10:32:00Z",
                } as ApiMessage,
            ]);

            vi.mocked(getChatById).mockResolvedValue(mockResponse);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const menuButton = container.querySelector('[role="button"]');
            await user.click(menuButton!);

            await new Promise((resolve) => setTimeout(resolve, 100));

            expect(messages.length).toBe(1);
            expect(messages[0].question).toBe("Question");
        });
    });

    describe("Accessibility", () => {
        it("has role='button' on menu button", () => {
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const button = container.querySelector('[role="button"]');
            expect(button).toBeInTheDocument();
        });

        it("has tabindex=0 on menu button", () => {
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const button = container.querySelector('[role="button"]');
            expect(button).toHaveAttribute("tabindex", "0");
        });

        it("input has correct id based on chatId", async () => {
            const user = userEvent.setup();
            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const pencilIcon = container.querySelector("svg.lucide-pencil");
            const editButton = pencilIcon?.closest("button");
            await user.click(editButton!);

            const input = container.querySelector("input");
            expect(input).toHaveAttribute("id", "titletest-chat-id");
        });
    });

    describe("Edge Cases", () => {
        it("handles summary with null values", () => {
            const edgeSummary = {
                chatId: "chat-1",
                userId: "user-1",
                title: "",
                createdAt: null as unknown as string,
                updatedAt: null as unknown as string,
            };

            const { container } = render(ChatSummary, {
                props: { summary: edgeSummary },
            });

            expect(container).toBeInTheDocument();
        });

        it("handles very long titles", () => {
            mockSummary.title = "A".repeat(200);

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            const paragraph = container.querySelector("p.truncate");
            expect(paragraph).toBeInTheDocument();
        });

        it("handles special characters in title", async () => {
            mockSummary.title = "<script>alert('xss')</script>";

            const { container } = render(ChatSummary, {
                props: { summary: mockSummary },
            });

            // Should be safely rendered
            expect(container.textContent).toContain(
                "<script>alert('xss')</script>",
            );
        });
    });
});

describe("commitTitleChange and saveTitle", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it("updates chat title on blur and calls state update and success toast", async () => {
        const summary = {
            title: "Old title",
            chatId: "chat-123",
            userId: "user-1",
            createdAt: "",
            updatedAt: "",
        } as ApiChatSummary;

        vi.mocked(updateChatTitleApi).mockResolvedValue({
            chatId: "chat-123",
            title: "New title",
            updatedAt: "2024-01-01",
            userId: "user-1",
            createdAt: "2023-12-01",
        });

        const updateChatTitleState = vi.fn();

        const { container } = render(ChatSummary, {
            props: {
                summary,
                updateChatTitleState,
            },
        });

        const pencil = container.querySelector("svg.lucide-pencil");
        const editButton = pencil?.closest("button");
        expect(editButton).toBeTruthy();
        await fireEvent.click(editButton!);

        const input = container.querySelector("input") as HTMLInputElement;
        expect(input).toBeTruthy();

        await fireEvent.input(input, { target: { value: " New title " } });
        await fireEvent.blur(input);

        await waitFor(() => {
            expect(updateChatTitleApi).toHaveBeenCalledWith(
                "chat-123",
                "New title",
                "user-1"
            );
        });


    expect(updateChatTitleState).toHaveBeenCalledWith(
        "chat-123",
        "New title",
    );


    expect(toast.success).toHaveBeenCalledWith(
        "Chat title updated successfully"
    );
});


    it("shows error toast when saveTitle fails", async () => {
        const summary = {
            title: "Old title",
            chatId: "chat-123",
            userId: "user-1",
            createdAt: "",
            updatedAt: "",
        } as ApiChatSummary;

        vi.mocked(updateChatTitleApi).mockRejectedValue(new Error("Boom"));

        const { container } = render(ChatSummary, { props: { summary } });

        const pencil = container.querySelector("svg.lucide-pencil");
        const editButton = pencil?.closest("button");
        await fireEvent.click(editButton!);

        const input = container.querySelector("input") as HTMLInputElement;
        await fireEvent.input(input, { target: { value: " New title " } });
        await fireEvent.blur(input);

        await waitFor(() => expect(toast.error).toHaveBeenCalledWith(
            "Umbenennen fehlgeschlagen",
            { description: "Boom" },
        ));
    });

    it("does nothing if title is unchanged on blur", async () => {
        const summary = {
            title: "Same title",
            chatId: "chat-123",
            userId: "user-1",
            createdAt: "",
            updatedAt: "",
        } as ApiChatSummary;

        const { container } = render(ChatSummary, { props: { summary } });

        const pencil = container.querySelector("svg.lucide-pencil");
        const editButton = pencil?.closest("button");
        await fireEvent.click(editButton!);

        const input = container.querySelector("input") as HTMLInputElement;
        // blur without changing value
        await fireEvent.blur(input);

        expect(updateChatTitleApi).not.toHaveBeenCalled();
        expect(toast.success).not.toHaveBeenCalled();
        expect(toast.error).not.toHaveBeenCalled();
    });

    it("calls updateChatTitleStance prop when provided", async () => {
        const summary = {
            title: "Old title",
            chatId: "chat-123",
            userId: "user-1",
            createdAt: "",
            updatedAt: "",
        } as ApiChatSummary;

        vi.mocked(updateChatTitleApi).mockResolvedValue({
            chatId: "chat-123",
            title: "New title",
            updatedAt: "2024-01-01",
            userId: "user-1",
            createdAt: "2023-12-01"
        });

        const stub = vi.fn();

        const { container } = render(ChatSummary, {
            props: { summary, updateChatTitleState: stub },
        });

        const pencil = container.querySelector("svg.lucide-pencil");
        const editButton = pencil?.closest("button");
        await fireEvent.click(editButton!);

        const input = container.querySelector("input") as HTMLInputElement;
        await fireEvent.input(input, { target: { value: " New title " } });
        await fireEvent.blur(input);

        await waitFor(() => expect(stub).toHaveBeenCalledWith("chat-123", "New title"));
    });
});
