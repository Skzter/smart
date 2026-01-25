import type { DateRange } from "bits-ui";

let updater: ((chatId: string, title: string) => void) | null = null;


export type MessageType = "user" | "validation" | "generation" | "error";
export type Message = { t: MessageType; Message: string };

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

export function registerChatTitleUpdater(
    fn: (chatId: string, title: string) => void,
) {
    updater = fn;
}

export function updateChatTitle(chatId: string, title: string) {
    updater?.(chatId, title);
}