import type { DateRange } from "bits-ui";

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
