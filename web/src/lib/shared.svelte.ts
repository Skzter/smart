import type { Message } from "$types/message";

export const user = $state({
    id: "",
});

export const chat = $state({
    id: "",
    isLoading: false,
});

export const messages: Message[] = $state([]);
