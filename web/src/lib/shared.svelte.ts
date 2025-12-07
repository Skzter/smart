import type { Message } from "$types/message";

export const messages: Message[] = $state([]);

export const user = $state({
    id: "",
});

export const chat = $state({
    id: "",
    isLoading: false,
});
