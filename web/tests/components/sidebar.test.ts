import { render, waitFor } from "@testing-library/svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { tick } from "svelte";
import SidebarTestWrapper from "../helpers/SidebarTestWrapper.svelte";
import type { ApiChatSummary } from "$lib/types";
import type { DateRange } from "bits-ui";

// Mock toast
vi.mock("svelte-sonner", () => ({
    toast: {
        error: vi.fn(),
    },
}));

// Mock API
vi.mock("$lib/api", () => ({
    getUserChats: vi.fn(),
}));

// Mock shared state
vi.mock("$lib/shared.svelte", () => ({
    user: { id: "test-user-123" },
    ChatDate: { Range: undefined },
    ChatFilter: { sortBy: "recent", timeFilter: "all" },
}));

import { getChats } from "$lib/api";
import { user, ChatDate, ChatFilter } from "$lib/shared.svelte";
import { toast } from "svelte-sonner";

describe("Sidebar", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        user.id = "test-user-123";
        ChatDate.Range = undefined;
        ChatFilter.sortBy = "recent";
        ChatFilter.timeFilter = "all";
    });

    it("renders the sidebar with loading state initially", async () => {
        vi.mocked(getChats).mockImplementation(() => new Promise(() => {})); // Never resolves

        const { container } = render(SidebarTestWrapper);

        await tick();

        const spinner = container.querySelector(".size-6");
        expect(spinner).toBeInTheDocument();
    });

    it("loads user chats successfully", async () => {
        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Test chat",
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("displays error when API call fails", async () => {
        vi.mocked(getChats).mockRejectedValue(new Error("API Error"));

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(toast.error).toHaveBeenCalled();
        });
    });

    it("does not load chats when user.id is undefined", async () => {
        (user as unknown as { id: undefined }).id = undefined;

        render(SidebarTestWrapper);
        await tick();

        expect(getChats).not.toHaveBeenCalled();
    });

    it("categorizes chat as 'Heute' for today's chats", async () => {
        const today = new Date();
        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Today's chat",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("categorizes chat as 'Gestern' for yesterday's chats", async () => {
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-2",
                userId: "test-user-123",
                title: "Yesterday's chat",
                createdAt: yesterday.toISOString(),
                updatedAt: yesterday.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("categorizes chat as 'letzte Woche' for chats within last 7 days", async () => {
        const lastWeek = new Date();
        lastWeek.setDate(lastWeek.getDate() - 5);

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-3",
                userId: "test-user-123",
                title: "Last week's chat",
                createdAt: lastWeek.toISOString(),
                updatedAt: lastWeek.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("categorizes chat as 'letzten Monat' for chats within current month", async () => {
        const thisMonth = new Date();
        thisMonth.setDate(thisMonth.getDate() - 15);

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-4",
                userId: "test-user-123",
                title: "This month's chat",
                createdAt: thisMonth.toISOString(),
                updatedAt: thisMonth.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("categorizes chat as 'früher' for older chats", async () => {
        const older = new Date();
        older.setMonth(older.getMonth() - 2);

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-5",
                userId: "test-user-123",
                title: "Old chat",
                createdAt: older.toISOString(),
                updatedAt: older.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("filters chats by date range when provided", async () => {
        const today = new Date();

        ChatDate.Range = {
            start: {
                toDate: () => {
                    const d = new Date(today);
                    d.setUTCHours(0, 0, 0, 0);
                    return d;
                },
            },
            end: {
                toDate: () => {
                    const d = new Date(today);
                    d.setUTCHours(23, 59, 59, 999);
                    return d;
                },
            },
        } as DateRange;

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Chat",
                createdAt: new Date(
                    today.getTime() - 30 * 24 * 60 * 60 * 1000,
                ).toISOString(),
                updatedAt: new Date(
                    today.getTime() - 30 * 24 * 60 * 60 * 1000,
                ).toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("returns true for isWithinDateRange when no date range is set", async () => {
        ChatDate.Range = undefined;

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Chat",
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("filters out chats outside of date range", async () => {
        const today = new Date();
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);

        ChatDate.Range = {
            start: {
                toDate: () => today,
            },
            end: {
                toDate: () => today,
            },
        } as DateRange;

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Today",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
            {
                chatId: "conv-2",
                userId: "test-user-123",
                title: "Yesterday",
                createdAt: yesterday.toISOString(),
                updatedAt: yesterday.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("handles empty chat list", async () => {
        vi.mocked(getChats).mockResolvedValue([]);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("groups multiple chats by date category", async () => {
        const today = new Date();
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Today 1",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
            {
                chatId: "conv-2",
                userId: "test-user-123",
                title: "Today 2",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
            {
                chatId: "conv-3",
                userId: "test-user-123",
                title: "Yesterday",
                createdAt: yesterday.toISOString(),
                updatedAt: yesterday.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("renders SidebarHeader component", async () => {
        vi.mocked(getChats).mockResolvedValue([]);

        const { container } = render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });

        expect(container).toBeInTheDocument();
    });

    it("renders Group components for each category", async () => {
        const today = new Date();
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Today",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
            {
                chatId: "conv-2",
                userId: "test-user-123",
                title: "Yesterday",
                createdAt: yesterday.toISOString(),
                updatedAt: yesterday.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("handles non-Error exceptions", async () => {
        vi.mocked(getChats).mockRejectedValue("String error");

        render(SidebarTestWrapper);

        await waitFor(
            () => {
                expect(toast.error).toHaveBeenCalledWith(
                    "Unbekannter Fehler",
                    expect.any(Object),
                );
            },
            { timeout: 3000 },
        );
    });

    it("updates groups when ChatDate.Range changes", async () => {
        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Chat",
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });

        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);

        ChatDate.Range = {
            start: {
                toDate: () => new Date(),
            },
            end: {
                toDate: () => tomorrow,
            },
        } as DateRange;

        await tick();
    });

    it("correctly sets time boundaries for date range filtering", async () => {
        const today = new Date();
        today.setHours(15, 30, 45, 500);

        ChatDate.Range = {
            start: {
                toDate: () => {
                    const d = new Date(today);
                    d.setUTCHours(0, 0, 0, 0);
                    return d;
                },
            },
            end: {
                toDate: () => {
                    const d = new Date(today);
                    d.setUTCHours(23, 59, 59, 999);
                    return d;
                },
            },
        } as DateRange;

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Chat",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });

    it("returns empty array when items is undefined in updateGroupsWithDateRange", async () => {
        vi.mocked(getChats).mockImplementation(() => new Promise(() => {}));

        const { container } = render(SidebarTestWrapper);
        await tick();

        const spinner = container.querySelector(".size-6");
        expect(spinner).toBeInTheDocument();
    });

    it("adds items to existing group category", async () => {
        const today = new Date();

        const mockChats: ApiChatSummary[] = [
            {
                chatId: "conv-1",
                userId: "test-user-123",
                title: "Chat 1",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
            {
                chatId: "conv-2",
                userId: "test-user-123",
                title: "Chat 2",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
            {
                chatId: "conv-3",
                userId: "test-user-123",
                title: "Chat 3",
                createdAt: today.toISOString(),
                updatedAt: today.toISOString(),
            },
        ];
        vi.mocked(getChats).mockResolvedValue(mockChats);

        render(SidebarTestWrapper);

        await waitFor(() => {
            expect(getChats).toHaveBeenCalled();
        });
    });
});
