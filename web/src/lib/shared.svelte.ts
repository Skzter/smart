import type { DateRange } from "bits-ui";

export type MessageType = "user" | "validation" | "generation" | "error";
export type Message = { t: MessageType; Message: string };
import type { ApiChatSummary } from "$types/api";

export const messages: Message[] = $state([]);

export const user = $state({
    id: "" as string | undefined,
});

export const chat = $state({
    id: "",
    isLoading: false,
});

export const ChatDate = $state({
    Range: undefined as DateRange | undefined,
});

export const ChatFilter = $state({
    sortBy: "recent" as "recent" | "created",
    timeFilter: "all" as "all" | "today" | "week" | "month",
});


let _chats = $state<ApiChatSummary[] | undefined>(undefined);

export function chats() {
    return _chats;
}

export function setChats(data: ApiChatSummary[]) {
    _chats = data;
}

/**
 * Update the title (and updatedAt) of a single chat.
 * Used after a successful PATCH /chats/:id.
 */
export function updateChatTitle(
    chatId: string,
    newTitle: string,
    updatedAt: string,
) {
    if (!_chats) return;

    _chats = _chats.map((chat) =>
        chat.chatId === chatId
            ? {
                  ...chat,
                  title: newTitle,
                  updatedAt,
              }
            : chat,
    );
}