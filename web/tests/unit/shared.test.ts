import { describe, it, expect, vi, beforeEach } from "vitest";
import {
    messages,
    user,
    chat,
    ChatDate,
    ChatFilter,
    registerChatTitleUpdater,
    updateChatTitle,
} from "$lib/shared.svelte";

describe("shared state", () => {
    beforeEach(() => {
        // reset mutable state manually
        messages.length = 0;

        user.id = "";

        chat.id = "";
        chat.isLoading = false;

        ChatDate.Range = undefined;

        ChatFilter.sortBy = "recent";
        ChatFilter.timeFilter = "all";
    });

    it("initializes messages as an empty array", () => {
        expect(Array.isArray(messages)).toBe(true);
        expect(messages.length).toBe(0);
    });

    it("allows pushing messages", () => {
        messages.push({ t: "user", Message: "Hello" });

        expect(messages.length).toBe(1);
        expect(messages[0]).toEqual({
            t: "user",
            Message: "Hello",
        });
    });

    it("initializes user state", () => {
        expect(user.id).toBe("");
    });

    it("updates user id", () => {
        user.id = "user-123";
        expect(user.id).toBe("user-123");
    });

    it("initializes chat state", () => {
        expect(chat.id).toBe("");
        expect(chat.isLoading).toBe(false);
    });

    it("updates chat state", () => {
        chat.id = "chat-456";
        chat.isLoading = true;

        expect(chat.id).toBe("chat-456");
        expect(chat.isLoading).toBe(true);
    });

    it("initializes ChatDate state", () => {
        expect(ChatDate.Range).toBeUndefined();
    });

    it("initializes ChatFilter state", () => {
        expect(ChatFilter.sortBy).toBe("recent");
        expect(ChatFilter.timeFilter).toBe("all");
    });
});

describe("chat title updater", () => {
    it("does nothing if no updater is registered", () => {
        // should not throw
        expect(() =>
            updateChatTitle("chat-id", "New Title"),
        ).not.toThrow();
    });

    it("calls registered updater with correct arguments", () => {
        const spy = vi.fn();

        registerChatTitleUpdater(spy);

        updateChatTitle("chat-123", "Generated Title");

        expect(spy).toHaveBeenCalledTimes(1);
        expect(spy).toHaveBeenCalledWith("chat-123", "Generated Title");
    });

    it("overwrites previous updater when registering a new one", () => {
        const first = vi.fn();
        const second = vi.fn();

        registerChatTitleUpdater(first);
        registerChatTitleUpdater(second);

        updateChatTitle("chat-999", "Title");

        expect(first).not.toHaveBeenCalled();
        expect(second).toHaveBeenCalledWith("chat-999", "Title");
    });
});
