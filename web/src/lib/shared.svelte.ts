import type { Message } from "$types/message";
import type { DateRange } from "bits-ui";

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
